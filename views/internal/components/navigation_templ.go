// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.543
package components

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import "context"
import "io"
import "bytes"

import (
	"github.com/MBvisti/grafto/server/middleware"
)

func extractAuthStatus(ctx context.Context) bool {
	if authCtx, ok := ctx.Value(middleware.AuthContext{}).(middleware.AuthContext); ok {
		return authCtx.GetAuthStatus()
	}

	return false
}

func Nav() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, templ_7745c5c3_W io.Writer) (templ_7745c5c3_Err error) {
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templ_7745c5c3_W.(*bytes.Buffer)
		if !templ_7745c5c3_IsBuffer {
			templ_7745c5c3_Buffer = templ.GetBuffer()
			defer templ.ReleaseBuffer(templ_7745c5c3_Buffer)
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<header class=\"container flex flex-wrap sm:justify-start sm:flex-nowrap z-50 w-full text-sm py-4 bg-gray-800\"><nav class=\"max-w-[85rem] w-full mx-auto px-4 sm:flex sm:items-center sm:justify-between\" aria-label=\"Global\"><div class=\"flex items-center justify-between\"><a class=\"flex-none text-xl font-semibold text-white\" href=\"/\">MBV</a><div class=\"sm:hidden\"><button type=\"button\" class=\"hs-collapse-toggle p-2 inline-flex justify-center items-center gap-x-2 rounded-lg border border-gray-200 bg-white text-gray-800 shadow-sm hover:bg-gray-50 disabled:opacity-50 disabled:pointer-events-none bg-transparent border-gray-700 text-white hover:bg-white/10 focus:outline-none focus:ring-1 focus:ring-gray-600\" data-hs-collapse=\"#navbar-collapse-with-animation\" aria-controls=\"navbar-collapse-with-animation\" aria-label=\"Toggle navigation\"><svg class=\"hs-collapse-open:hidden flex-shrink-0 w-4 h-4\" xmlns=\"http://www.w3.org/2000/svg\" width=\"24\" height=\"24\" viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><line x1=\"3\" x2=\"21\" y1=\"6\" y2=\"6\"></line><line x1=\"3\" x2=\"21\" y1=\"12\" y2=\"12\"></line><line x1=\"3\" x2=\"21\" y1=\"18\" y2=\"18\"></line></svg> <svg class=\"hs-collapse-open:block hidden flex-shrink-0 w-4 h-4\" xmlns=\"http://www.w3.org/2000/svg\" width=\"24\" height=\"24\" viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><path d=\"M18 6 6 18\"></path><path d=\"m6 6 12 12\"></path></svg></button></div></div><div id=\"navbar-collapse-with-animation\" class=\"hs-collapse hidden overflow-hidden transition-all duration-300 basis-full grow sm:block\"><div class=\"flex flex-col gap-5 mt-5 sm:flex-row sm:items-center sm:justify-end sm:mt-0 sm:ps-5\"><a class=\"font-medium text-blue-500 focus:outline-none focus:ring-1 focus:ring-gray-600\" href=\"/\">Home</a> <a class=\"font-medium hover:text-gray-400 text-gray-400 hover:text-gray-500 focus:outline-none focus:ring-1 focus:ring-gray-600\" href=\"/newsletter\">Newsletter</a> <a class=\"font-medium hover:text-gray-400 text-gray-400 hover:text-gray-500 focus:outline-none focus:ring-1 focus:ring-gray-600\" href=\"/projects\">Projects</a> <a class=\"font-medium hover:text-gray-400 text-gray-400 hover:text-gray-500 focus:outline-none focus:ring-1 focus:ring-gray-600\" href=\"/about\">About</a> ")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if extractAuthStatus(ctx) {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<a class=\"font-medium text-gray-400 hover:text-gray-500\" href=\"/user/logout\">logout</a>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</div></div></nav></header>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if !templ_7745c5c3_IsBuffer {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteTo(templ_7745c5c3_W)
		}
		return templ_7745c5c3_Err
	})
}
