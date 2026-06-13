package handler

// contactForm carries submitted values and per-field validation errors
// back into the form partial.
type contactForm struct {
	Name    string
	Email   string
	Company string
	Kind    string
	Message string
	Errors  map[string]string
}
