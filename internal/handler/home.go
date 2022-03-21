package handler

import (
	"log"
	"net/http"

	"github.com/petaki/inertia-go"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	im := ctx.Value("inertia").(*inertia.Inertia)
	err := im.Render(w, r, "Welcome", nil)
	if err != nil {
		log.Panic(err)
	}
}
