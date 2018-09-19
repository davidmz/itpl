/*
Package itpl allows to include one Go templates into anothers.

This package introduce an 'include' action that works like that:
	// "page.tpl"
	{{include "header.tpl"}}
	Content
	{{include "footer.tpl"}}

	// "header.tpl"
	Header

	// "footer.tpl"
	Footer

itpl.Load("page.tpl") will load "page.tpl" file, include content of "header.tpl" and "footer.tpl" files
at the places of 'include' actions and produce the following template code:
	Header
	Content
	Footer

Of course any more complex template logic can be used. The Load function returns a combined template as a string
that can be parsed and executed used with text/template of html/template package.

The include action required one string argument, the relative or absolute path to included file. If relative
path is used then it will be resolved relative to file with the include action:
	// /1.tpl
	{{include "inc/2.tpl"}}

	// /inc/2.tpl
	{{include "3.tpl"}}

	// /inc/3.tpl
	Some content

This package does not execute templates, it just parses them and combine content
of multiple files into one. So the include actions cannot use any dynamic variables
or parameters to construct the file path. File path must be a regular quoted string.
*/
package itpl
