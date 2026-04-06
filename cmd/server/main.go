package main

import (
	"fmt"
	"log"
	"net/http"

	"chat-go/config"
	"chat-go/internal/cache"
	"chat-go/internal/database"
	"chat-go/internal/handler"
	"chat-go/internal/middleware"
	"chat-go/internal/repository"
	"chat-go/internal/service"
	"chat-go/internal/websocket"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	db, err := database.NewConnection(cfg.DB)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Connect to Redis
	redisClient, err := cache.NewRedisClient(
		cfg.Redis.Host,
		cfg.Redis.Port,
		cfg.Redis.Password,
		cfg.Redis.DB,
	)
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
		log.Println("Continuing without Redis caching...")
		redisClient = nil
	}
	if redisClient != nil {
		defer redisClient.Close()
	}

	// Initialize cache (nil-safe, will use in-memory fallbacks if Redis unavailable)
	var appCache *cache.Cache
	if redisClient != nil {
		appCache = cache.NewCache(redisClient)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	roomRepo := repository.NewRoomRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	notifRepo := repository.NewNotificationRepository(db)

	// Initialize WebSocket hub
	hub := websocket.NewHub()
	go hub.Run()

	// Initialize services
	authService := service.NewAuthService(userRepo, cfg.JWT)

	// Initialize notification service with WebSocket notifier
	// VAPID keys would come from config in production
	var vapidKeys *service.VAPIDKeys
	if cfg.VAPID != nil && cfg.VAPID.PublicKey != "" {
		vapidKeys = &service.VAPIDKeys{
			PublicKey:  cfg.VAPID.PublicKey,
			PrivateKey: cfg.VAPID.PrivateKey,
			Subject:    cfg.VAPID.Subject,
		}
	}
	notifService := service.NewNotificationService(notifRepo, hub, vapidKeys)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userRepo)
	roomHandler := handler.NewRoomHandler(roomRepo)
	messageHandler := handler.NewMessageHandler(messageRepo, roomRepo, appCache)
	notifHandler := handler.NewNotificationHandler(notifService)
	wsHandler := handler.NewWebSocketHandler(hub, authService, userRepo, messageRepo, appCache)

	// Create router
	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("/api/auth/register", authHandler.Register)
	mux.HandleFunc("/api/auth/login", authHandler.Login)

	// WebSocket endpoint
	mux.HandleFunc("/ws", wsHandler.ServeWS)

	// Protected routes - wrap with auth middleware
	protectedMux := http.NewServeMux()

	// User routes
	protectedMux.HandleFunc("/api/users", userHandler.GetUsers)
	protectedMux.HandleFunc("/api/users/", userHandler.GetUser)
	protectedMux.HandleFunc("/api/me", userHandler.GetMe)

	// Room routes
	protectedMux.HandleFunc("/api/rooms", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			roomHandler.GetRooms(w, r)
		case http.MethodPost:
			roomHandler.CreateRoom(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	protectedMux.HandleFunc("/api/rooms/my", roomHandler.GetMyRooms)
	protectedMux.HandleFunc("/api/rooms/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if len(path) > 11 && path[len(path)-5:] == "/join" {
			roomHandler.JoinRoom(w, r)
		} else if len(path) > 12 && path[len(path)-6:] == "/leave" {
			roomHandler.LeaveRoom(w, r)
		} else if len(path) > 14 && path[len(path)-9:] == "/messages" {
			switch r.Method {
			case http.MethodGet:
				messageHandler.GetRoomMessages(w, r)
			case http.MethodPost:
				messageHandler.SendRoomMessage(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else if len(path) > 18 && path[len(path)-14:] == "/messages/read" {
			messageHandler.MarkRoomMessagesAsRead(w, r)
		} else if len(path) > 20 && path[len(path)-16:] == "/messages/unread" {
			messageHandler.GetUnreadRoomMessagesCount(w, r)
		} else {
			roomHandler.GetRoom(w, r)
		}
	})

	// Message edit/delete routes
	protectedMux.HandleFunc("/api/messages/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut, http.MethodPatch:
			messageHandler.EditRoomMessage(w, r)
		case http.MethodDelete:
			messageHandler.DeleteRoomMessage(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Direct messages routes
	protectedMux.HandleFunc("/api/dm/unread", messageHandler.GetUnreadDirectMessagesCount)
	protectedMux.HandleFunc("/api/dm/messages/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut, http.MethodPatch:
			messageHandler.EditDirectMessage(w, r)
		case http.MethodDelete:
			messageHandler.DeleteDirectMessage(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	protectedMux.HandleFunc("/api/dm/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			messageHandler.GetDirectMessages(w, r)
		case http.MethodPost:
			messageHandler.SendDirectMessage(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Combined chat list routes (both DM and rooms)
	protectedMux.HandleFunc("/api/chats", messageHandler.GetChatList)
	protectedMux.HandleFunc("/api/chats/", messageHandler.GetChat)

	// Notification routes
	protectedMux.HandleFunc("/api/notifications", func(w http.ResponseWriter, r *http.Request) {
		notifHandler.GetNotifications(w, r)
	})
	protectedMux.HandleFunc("/api/notifications/count", notifHandler.GetUnreadCount)
	protectedMux.HandleFunc("/api/notifications/counts", notifHandler.GetUnreadCounts)
	protectedMux.HandleFunc("/api/notifications/read-all", notifHandler.MarkAllAsRead)
	protectedMux.HandleFunc("/api/notifications/preferences", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			notifHandler.GetPreferences(w, r)
		case http.MethodPut, http.MethodPatch:
			notifHandler.UpdatePreferences(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	protectedMux.HandleFunc("/api/notifications/push/subscribe", notifHandler.RegisterPushSubscription)
	protectedMux.HandleFunc("/api/notifications/push/unsubscribe", notifHandler.UnregisterPushSubscription)
	protectedMux.HandleFunc("/api/notifications/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if len(path) > 22 && path[len(path)-5:] == "/read" {
			notifHandler.MarkAsRead(w, r)
		} else if r.Method == http.MethodDelete {
			notifHandler.DeleteNotification(w, r)
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	})

	// Apply auth middleware to protected routes
	authMiddleware := middleware.AuthMiddleware(authService)
	mux.Handle("/api/users", authMiddleware(protectedMux))
	mux.Handle("/api/users/", authMiddleware(protectedMux))
	mux.Handle("/api/me", authMiddleware(protectedMux))
	mux.Handle("/api/rooms", authMiddleware(protectedMux))
	mux.Handle("/api/rooms/", authMiddleware(protectedMux))
	mux.Handle("/api/messages/", authMiddleware(protectedMux))
	mux.Handle("/api/chats", authMiddleware(protectedMux))
	mux.Handle("/api/chats/", authMiddleware(protectedMux))
	mux.Handle("/api/dm/", authMiddleware(protectedMux))
	mux.Handle("/api/dm/unread", authMiddleware(protectedMux))
	mux.Handle("/api/notifications", authMiddleware(protectedMux))
	mux.Handle("/api/notifications/", authMiddleware(protectedMux))

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	})

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/", fs)

	// Apply global middleware
	finalHandler := middleware.Logging(middleware.CORS(mux))

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	fmt.Printf("🚀 Server starting on http://localhost%s\n", addr)
	fmt.Println("📡 WebSocket endpoint: ws://localhost" + addr + "/ws")
	fmt.Println("🌐 Frontend: http://localhost" + addr + "/")
	fmt.Println("🔔 Notifications enabled")
	if appCache != nil {
		fmt.Println("📦 Redis caching enabled")
	} else {
		fmt.Println("⚠️  Redis caching disabled (connection failed)")
	}

	if err := http.ListenAndServe(addr, finalHandler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
