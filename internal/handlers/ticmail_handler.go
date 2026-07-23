package handlers

import (
	"html/template"
	"net/http"
	"ticDesk/internal/auth"
	"ticDesk/internal/repository"
)

type TicMailHandler struct {
	ticmailRepo *repository.TicMailRepository
	tmpl        *template.Template
}

func NewTicMailHandler(ticmailRepo *repository.TicMailRepository, tmpl *template.Template) *TicMailHandler {
	return &TicMailHandler{
		ticmailRepo: ticmailRepo,
		tmpl:        tmpl,
	}
}

func (h *TicMailHandler) RenderTicMail(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	logs, err := h.ticmailRepo.ListLogs(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"User": user,
		"Logs": logs,
	}
	h.tmpl.ExecuteTemplate(w, "ticmail.html", data)
}

func (h *TicMailHandler) ClearTicMail(w http.ResponseWriter, r *http.Request) {
	_ = h.ticmailRepo.ClearLogs(r.Context())
	http.Redirect(w, r, "/ticmail", http.StatusSeeOther)
}
