// Code generated by templ - DO NOT EDIT.

// templ: version: v0.3.833
package account

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

import "gosl/pkg/contexts"

func ChangeUsername(err string, username string) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)

		user := contexts.GetUser(ctx)
		if username == "" {
			username = user.Username
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 1, "<form hx-post=\"/change-username\" hx-swap=\"outerHTML\" class=\"w-[90%] mx-auto mt-5\" x-data=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var2 string
		templ_7745c5c3_Var2, templ_7745c5c3_Err = templ.JoinStringErrs(templ.JSFuncCall(
			"usernameComponent", username, user.Username, err,
		).CallInline)
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `internal/view/component/account/changeusername.templ`, Line: 18, Col: 32}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var2))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 2, "\"><script>\n            function usernameComponent(newUsername, oldUsername, err) {\n                return {\n                    username: newUsername,\n                    initialUsername: oldUsername, \n                    err: err,\n                    resetUsername() {\n                        this.username = this.initialUsername;\n                        this.err = \"\";\n                    },\n                };\n            }\n        </script><div class=\"flex flex-col sm:flex-row\"><div class=\"flex flex-col sm:flex-row sm:items-center relative\"><label for=\"username\" class=\"text-lg w-20\">Username</label> <input type=\"text\" id=\"username\" name=\"username\" class=\"py-1 px-4 rounded-lg text-md\n                    bg-surface0 border border-surface2 w-50 sm:ml-5\n                    disabled:opacity-50 ml-0 disabled:pointer-events-none\" required aria-describedby=\"username-error\" x-model=\"username\"><div class=\"absolute inset-y-0 sm:start-68 start-43 pt-9\n                        pointer-events-none sm:pt-2\" x-show=\"err\" x-cloak><svg class=\"size-5 text-red\" width=\"16\" height=\"16\" fill=\"currentColor\" viewBox=\"0 0 16 16\" aria-hidden=\"true\"><path d=\"M16 8A8 8 0 1 1 0 8a8 8 0 0 1 16 0zM8 \n                                4a.905.905 0 0 0-.9.995l.35 3.507a.552.552 0 0 \n                                0 1.1 0l.35-3.507A.905.905 0 0 0 8 4zm.002 6a1 \n                                1 0 1 0 0 2 1 1 0 0 0 0-2z\"></path></svg></div></div><div class=\"mt-2 sm:mt-0\"><button class=\"rounded-lg bg-blue py-1 px-2 text-mantle sm:ml-2\n                hover:cursor-pointer hover:bg-blue/75 transition\" x-cloak x-show=\"username !== initialUsername\" x-transition.opacity.duration.500ms>Update</button> <button class=\"rounded-lg bg-overlay0 py-1 px-2 text-mantle\n                hover:cursor-pointer hover:bg-surface2 transition\" type=\"button\" href=\"#\" x-cloak x-show=\"username !== initialUsername\" x-transition.opacity.duration.500ms @click=\"resetUsername()\">Cancel</button></div></div><p class=\"block text-red sm:ml-26 mt-1 transition\" x-cloak x-show=\"err\" x-text=\"err\"></p></form>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return nil
	})
}

var _ = templruntime.GeneratedTemplate
