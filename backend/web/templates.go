package web

import (
	"html/template"
	"io/fs"
)

const (
	profileShowKey = "profile_show"
	signUpKey      = "signup"
	loginKey       = "login"
	header         = "templates/layout/header.gohtml"
	footer         = "templates/layout/footer.gohtml"
)

var pageToTemplate = map[string]string{
	profileShowKey: "templates/profiles/show.gohtml",
	signUpKey:      "templates/sessions/signup.gohtml",
	loginKey:       "templates/sessions/login.gohtml",
}

func BuildTemplates(templateFS fs.FS) map[string]*template.Template {
	tmpls := make(map[string]*template.Template)

	for key, tmplFile := range pageToTemplate {
		tmpls[key] = template.Must(template.ParseFS(templateFS, "templates/layout/*.gohtml", tmplFile))
	}
	return tmpls
}
