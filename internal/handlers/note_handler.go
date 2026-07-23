package handlers

import (
	"html/template"
	"net/http"
	"ticDesk/internal/auth"
	"ticDesk/internal/models"
	"ticDesk/internal/repository"
)

type NoteHandler struct {
	noteRepo *repository.NoteRepository
	tmpl     *template.Template
}

func NewNoteHandler(noteRepo *repository.NoteRepository, tmpl *template.Template) *NoteHandler {
	return &NoteHandler{
		noteRepo: noteRepo,
		tmpl:     tmpl,
	}
}

func (h *NoteHandler) RenderNotes(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	notes, err := h.noteRepo.ListNotes(r.Context(), user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Build a separate slice of pinned notes for the template's pinned section
	var pinnedNotes []models.UserNote
	for _, n := range notes {
		if n.IsPinned {
			pinnedNotes = append(pinnedNotes, n)
		}
	}

	data := map[string]interface{}{
		"User":        user,
		"Notes":       notes,
		"PinnedNotes": pinnedNotes,
	}
	h.tmpl.ExecuteTemplate(w, "notes.html", data)
}

func (h *NoteHandler) CreateNote(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")
	color := r.FormValue("color")
	isPinned := r.FormValue("is_pinned") == "true"

	if title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}
	if color == "" {
		color = "blue"
	}

	_, err := h.noteRepo.CreateNote(r.Context(), user.ID, title, content, color, isPinned)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/notes", http.StatusSeeOther)
}

func (h *NoteHandler) TogglePin(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if id != "" {
		_ = h.noteRepo.TogglePin(r.Context(), id, user.ID)
	}
	http.Redirect(w, r, "/notes", http.StatusSeeOther)
}

func (h *NoteHandler) DeleteNote(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if id != "" {
		_ = h.noteRepo.DeleteNote(r.Context(), id, user.ID)
	}
	http.Redirect(w, r, "/notes", http.StatusSeeOther)
}
