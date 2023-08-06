package http

import "html/template"

const (
	profileShowKey = "profile_show"
	header         = "./templates/layout/header.gohtml"
	footer         = "./templates/layout/footer.gohtml"
)

var pageToTemplate = map[string]string{
	profileShowKey: "./templates/profiles/show.gohtml",
}

func BuildTemplates() map[string]*template.Template {
	tmpls := make(map[string]*template.Template)
	for key, tmplFile := range pageToTemplate {
		tmpls[key] = template.Must(template.ParseFiles(header, footer, tmplFile))
	}
	return tmpls
}
