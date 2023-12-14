package main

import (
    "encoding/base64"
    "fmt"
    "io"
    "net/http"
    "os"
    "strings"
)

// Replace these with your desired credentials
const Username = "admin"
const Password = "harness"

func basicAuth(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        auth := r.Header.Get("Authorization")
        if auth == "" {
            w.Header().Set("WWW-Authenticate", `Basic realm="restricted"`)
            http.Error(w, "authorization required", http.StatusUnauthorized)
            return
        }

        payload, _ := base64.StdEncoding.DecodeString(strings.TrimPrefix(auth, "Basic "))
        pair := strings.SplitN(string(payload), ":", 2)

        if len(pair) != 2 || !validateCredentials(pair[0], pair[1]) {
            http.Error(w, "authorization failed", http.StatusUnauthorized)
            return
        }

        next(w, r)
    }
}

func validateCredentials(username, password string) bool {
    return username == Username && password == Password
}

func uploadFileHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Parse the multipart form
        if err := r.ParseMultipartForm(10 << 20); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        // Retrieve the file from form data
        file, handler, err := r.FormFile("file")
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
        defer file.Close()

        // Specify the directory and filename where the file will be saved
        filePath := "./uploads/" + handler.Filename

        // Create the file
        dst, err := os.Create(filePath)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        defer dst.Close()

        // Copy the uploaded file to the filesystem at the specified destination
        if _, err := io.Copy(dst, file); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        fmt.Fprintf(w, "File uploaded successfully: %s", handler.Filename)
    }
}

func main() {
    http.HandleFunc("/upload", basicAuth(uploadFileHandler()))

    // Start the server
    fmt.Println("Server started on :8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        fmt.Println("Server failed to start:", err)
    }
}
