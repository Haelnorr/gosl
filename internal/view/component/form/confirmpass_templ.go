// Code generated by templ - DO NOT EDIT.

// templ: version: v0.3.833
package form

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

func ConfirmPassword(err string) templ.Component {
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
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 1, "<form hx-post=\"/reauthenticate\" x-data=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var2 string
		templ_7745c5c3_Var2, templ_7745c5c3_Err = templ.JoinStringErrs(templ.JSFuncCall(
			"confirmPassData", err,
		).CallInline)
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `internal/view/component/form/confirmpass.templ`, Line: 8, Col: 28}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var2))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 2, "\" x-on:htmx:xhr:loadstart=\"submitted=true;buttontext=&#39;Loading...&#39;\"><script>\n            function confirmPassData(err) {\n                return {\n                    submitted: false,\n                    buttontext: 'Confirm', \n                    errMsg: err,\n                    reset() {\n                        this.err = \"\";\n                    },\n                };\n            }\n        </script><div class=\"grid gap-y-4\"><div class=\"mt-5\"><div class=\"relative\"><input type=\"password\" id=\"password\" name=\"password\" class=\"py-3 px-4 block w-full rounded-lg text-sm\n                        focus:border-blue focus:ring-blue bg-base\n                        disabled:opacity-50 disabled:pointer-events-none\" placeholder=\"Confirm password\" required aria-describedby=\"password-error\" @input=\"reset()\"><div class=\"absolute inset-y-0 end-0 \n                        pointer-events-none pe-3 pt-3\" x-show=\"errMsg\" x-cloak><svg class=\"size-5 text-red\" width=\"16\" height=\"16\" fill=\"currentColor\" viewBox=\"0 0 16 16\" aria-hidden=\"true\"><path d=\"M16 8A8 8 0 1 1 0 8a8 8 0 0 1 16 0zM8 \n                                4a.905.905 0 0 0-.9.995l.35 3.507a.552.552 0 0\n                                0 1.1 0l.35-3.507A.905.905 0 0 0 8 4zm.002 6a1\n                                1 0 1 0 0 2 1 1 0 0 0 0-2z\"></path></svg></div></div><p class=\"text-center text-xs text-red mt-2\" id=\"password-error\" x-show=\"errMsg\" x-cloak x-text=\"errMsg\"></p></div><button x-bind:disabled=\"submitted\" x-text=\"buttontext\" type=\"submit\" class=\"w-full py-3 px-4 inline-flex justify-center items-center \n                gap-x-2 rounded-lg border border-transparent transition\n                bg-blue hover:bg-blue/75 text-mantle hover:cursor-pointer\n                disabled:bg-blue/60 disabled:cursor-default\"></button> <button type=\"button\" class=\"w-full py-3 px-4 inline-flex justify-center items-center \n                gap-x-2 rounded-lg border border-transparent transition\n                bg-surface2 hover:bg-surface1 hover:cursor-pointer\n                disabled:cursor-default\" @click=\"showConfirmPasswordModal=false\">Cancel</button></div></form>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return nil
	})
}

var _ = templruntime.GeneratedTemplate
