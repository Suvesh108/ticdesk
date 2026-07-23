package handlers

import (
	"html/template"
	"net/http"
	"ticDesk/internal/auth"
	"ticDesk/internal/repository"
	"time"
)

type CalendarHandler struct {
	scheduleRepo *repository.ScheduleRepository
	tmpl         *template.Template
}

func NewCalendarHandler(scheduleRepo *repository.ScheduleRepository, tmpl *template.Template) *CalendarHandler {
	return &CalendarHandler{
		scheduleRepo: scheduleRepo,
		tmpl:         tmpl,
	}
}

func (h *CalendarHandler) RenderCalendar(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	events, err := h.scheduleRepo.ListEvents(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.New("").Funcs(template.FuncMap{
		"sub": func(a, b int) int { return a - b },
		"add": func(a, b int) int { return a + b },
		"mul": func(a, b int) int { return a * b },
	}).ParseFiles(
		"web/templates/layouts/base.html",
		"web/templates/pages/calendar.html",
	)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title":  "Schedule Calendar — ticDesk",
		"User":   user,
		"Events": events,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.ExecuteTemplate(w, "base.html", data)
}

func (h *CalendarHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")
	eventType := r.FormValue("event_type")
	startStr := r.FormValue("start_time")
	endStr := r.FormValue("end_time")

	if title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	startTime, err := time.Parse("2006-01-02T15:04", startStr)
	if err != nil {
		startTime = time.Now()
	}

	endTime, err := time.Parse("2006-01-02T15:04", endStr)
	if err != nil {
		endTime = startTime.Add(1 * time.Hour)
	}

	if eventType == "" {
		eventType = "maintenance"
	}

	_, err = h.scheduleRepo.CreateEvent(r.Context(), title, description, eventType, startTime, endTime, user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/calendar", http.StatusSeeOther)
}

func (h *CalendarHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id != "" {
		_ = h.scheduleRepo.DeleteEvent(r.Context(), id)
	}
	http.Redirect(w, r, "/calendar", http.StatusSeeOther)
}
