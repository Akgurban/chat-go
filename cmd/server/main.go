package main

import (
	"fmt"
	"log"
	"net/http"

	"chat-go/config"
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

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	roomRepo := repository.NewRoomRepository(db)
	messageRepo := repository.NewMessageRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo, cfg.JWT)

	// Initialize WebSocket hub
	hub := websocket.NewHub()
	go hub.Run()

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userRepo)
	roomHandler := handler.NewRoomHandler(roomRepo)
	messageHandler := handler.NewMessageHandler(messageRepo, roomRepo)
	wsHandler := handler.NewWebSocketHandler(hub, authService, userRepo, messageRepo)

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
		} else {
			roomHandler.GetRoom(w, r)
		}
	})

	// Direct messages routes
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

	// Apply auth middleware to protected routes
	authMiddleware := middleware.AuthMiddleware(authService)
	mux.Handle("/api/users", authMiddleware(protectedMux))
	mux.Handle("/api/users/", authMiddleware(protectedMux))
	mux.Handle("/api/me", authMiddleware(protectedMux))
	mux.Handle("/api/rooms", authMiddleware(protectedMux))
	mux.Handle("/api/rooms/", authMiddleware(protectedMux))
	mux.Handle("/api/dm/", authMiddleware(protectedMux))

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

	if err := http.ListenAndServe(addr, finalHandler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
