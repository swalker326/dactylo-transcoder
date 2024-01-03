package upload

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}
	r.ParseMultipartForm(10 << 20) // 10 MB max
	file, handler, err := r.FormFile("file")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}
	sign := r.FormValue("sign")
	defer file.Close()

	// You can get the content type from the FileHeader
	contentType := handler.Header.Get("Content-Type")

	// Upload to R2
	response, err := UploadToStorage(file, sign, contentType)
	if err != nil {
		fmt.Fprintf(w, "Failed to upload: %s\n", err)
		return
	}

	// Return the URL of the uploaded file
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}
