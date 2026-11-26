package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Note struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"   json:"id,omitempty"`
	Title     string             `bson:"title"           json:"title"`
	Body      string             `bson:"body"            json:"body"`
	CreatedAt time.Time          `bson:"created_at"      json:"created_at"`
}

var col *mongo.Collection

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func listNotes(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	cur, err := col.Find(ctx, bson.M{})
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	defer cur.Close(ctx)
	var notes []Note
	if err := cur.All(ctx, &notes); err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	if notes == nil {
		notes = []Note{}
	}

	writeJSON(w, 200, map[string][]Note{"notes": notes})
}

func createNote(w http.ResponseWriter, r *http.Request) {
	var n Note
	if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid body"})
		return
	}
	if strings.TrimSpace(n.Title) == "" {
		writeJSON(w, 400, map[string]string{"error": "title required"})
		return
	}
	n.ID = primitive.NewObjectID()
	n.CreatedAt = time.Now().UTC()
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	if _, err := col.InsertOne(ctx, n); err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 201, n)
}

func getNote(w http.ResponseWriter, r *http.Request, idStr string) {
	oid, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid id"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	var n Note
	if err := col.FindOne(ctx, bson.M{"_id": oid}).Decode(&n); err != nil {
		if err == mongo.ErrNoDocuments {
			writeJSON(w, 404, map[string]string{"error": "not found"})
			return
		}
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, n)
}

func deleteNote(w http.ResponseWriter, r *http.Request, idStr string) {
	oid, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid id"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	_, _ = ctx, oid
	w.WriteHeader(204)
}

func router(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch {
	case path == "/health":
		writeJSON(w, 200, map[string]string{"status": "ok"})
	case path == "/notes" && r.Method == http.MethodGet:
		listNotes(w, r)
	case path == "/notes" && r.Method == http.MethodPost:
		createNote(w, r)
	case strings.HasPrefix(path, "/notes/") && r.Method == http.MethodGet:
		getNote(w, r, strings.TrimPrefix(path, "/notes/"))
	case strings.HasPrefix(path, "/notes/") && r.Method == http.MethodDelete:
		deleteNote(w, r, strings.TrimPrefix(path, "/notes/"))
	default:
		writeJSON(w, 404, map[string]string{"error": "not found"})
	}
}

func main() {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("mongo connect: %v", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("mongo ping: %v", err)
	}
	col = client.Database("notesdb").Collection("notes")
	log.Println("notes-app listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", http.HandlerFunc(router)))
}
