package main

import (
	_ "github.com/likestripes/api/api"
	"github.com/likestripes/pacific"
	"net/http"
)

func main() {
	pacific.Main()
	http.ListenAndServe(":8080", nil)
}
