// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.778
package sections

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

func Features() templ.Component {
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
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<!-- Features #3 --><section class=\"relative bg-zinc-800 after:absolute after:top-0 after:right-0 after:h-full after:w-96 after:pointer-events-none after:bg-gradient-to-l after:from-zinc-800 max-lg:after:hidden\"><div class=\"py-12 md:py-20\"><!-- Carousel --><div class=\"max-w-xl lg:max-w-6xl mx-auto px-8 sm:px-6\"><div class=\"lg:flex space-y-12 lg:space-y-0 lg:space-x-12 xl:space-x-24\" x-data=\"{ tab: &#39;1&#39; }\"><!-- Content --><div class=\"lg:max-w-none lg:min-w-[524px]\"><div class=\"mb-8\"><div class=\"inline-flex text-sm font-medium text-zinc-400 px-4 py-0.5 border border-transparent [background:linear-gradient(theme(colors.zinc.800),theme(colors.zinc.800))_padding-box,linear-gradient(120deg,theme(colors.zinc.700),theme(colors.zinc.700/0),theme(colors.zinc.700))_border-box] rounded-full mb-4\">L1 Blockchain</div><h3 class=\"font-inter-tight text-3xl font-bold text-zinc-200 mb-4\">The Protocol for Decentralized Identity & Authentication</h3><p class=\"text-lg text-zinc-500\">We're creating the Global Standard for Decentralized Identity.  Authenticate users with PassKeys, Issue Crypto Wallets, Build  Payment flows, Send Encrypted Messages - all on a single platform.</p></div><!-- Tabs buttons --><div class=\"mb-8 md:mb-0 space-y-2\"><button :class=\"tab !== &#39;1&#39; ? &#39;&#39; : &#39;[background:linear-gradient(#2E2E32,#2E2E32)_padding-box,linear-gradient(120deg,theme(colors.zinc.700),theme(colors.zinc.700/0),theme(colors.zinc.700))_border-box]&#39;\" class=\"text-left flex items-center px-6 py-4 rounded border border-transparent\" @click.prevent=\"tab = &#39;1&#39;\"><svg class=\"shrink-0 fill-zinc-400 mr-3\" xmlns=\"http://www.w3.org/2000/svg\" width=\"24\" height=\"24\"><path d=\"m7.951 14.537 6.296-7.196 1.506 1.318-7.704 8.804-3.756-3.756 1.414-1.414 2.244 2.244Zm11.296-7.196 1.506 1.318-7.704 8.804-1.756-1.756 1.414-1.414.244.244 6.296-7.196Z\"></path></svg><div><div class=\"font-inter-tight text-lg font-semibold text-zinc-200 mb-1\">Make designs feel real</div><div class=\"text-zinc-500\">Save time and keep things consistent with reusable images, and 3D assets in shared libraries.</div></div></button> <button :class=\"tab !== &#39;2&#39; ? &#39;&#39; : &#39;[background:linear-gradient(#2E2E32,#2E2E32)_padding-box,linear-gradient(120deg,theme(colors.zinc.700),theme(colors.zinc.700/0),theme(colors.zinc.700))_border-box]&#39;\" class=\"text-left flex items-center px-6 py-4 rounded border border-transparent\" @click.prevent=\"tab = &#39;2&#39;\"><svg class=\"shrink-0 fill-zinc-400 mr-3\" xmlns=\"http://www.w3.org/2000/svg\" width=\"24\" height=\"24\"><path d=\"m16.997 19.056-1.78-.912A13.91 13.91 0 0 0 16.75 11.8c0-2.206-.526-4.38-1.533-6.344l1.78-.912A15.91 15.91 0 0 1 18.75 11.8c0 2.524-.602 5.01-1.753 7.256Zm-3.616-1.701-1.77-.93A9.944 9.944 0 0 0 12.75 11.8c0-1.611-.39-3.199-1.14-4.625l1.771-.93c.9 1.714 1.37 3.62 1.369 5.555 0 1.935-.47 3.841-1.369 5.555Zm-3.626-1.693-1.75-.968c.49-.885.746-1.881.745-2.895a5.97 5.97 0 0 0-.745-2.893l1.75-.968a7.968 7.968 0 0 1 .995 3.861 7.97 7.97 0 0 1-.995 3.863Zm-3.673-1.65-1.664-1.11c.217-.325.333-.709.332-1.103 0-.392-.115-.776-.332-1.102L6.082 9.59c.437.655.67 1.425.668 2.21a3.981 3.981 0 0 1-.668 2.212Z\"></path></svg><div><div class=\"font-inter-tight text-lg font-semibold text-zinc-200 mb-1\">Bring creatives closer</div><div class=\"text-zinc-500\">Save time and keep things consistent with reusable images, and 3D assets in shared libraries.</div></div></button> <button :class=\"tab !== &#39;3&#39; ? &#39;&#39; : &#39;[background:linear-gradient(#2E2E32,#2E2E32)_padding-box,linear-gradient(120deg,theme(colors.zinc.700),theme(colors.zinc.700/0),theme(colors.zinc.700))_border-box]&#39;\" class=\"text-left flex items-center px-6 py-4 rounded border border-transparent\" @click.prevent=\"tab = &#39;3&#39;\"><svg class=\"shrink-0 fill-zinc-400 mr-3\" xmlns=\"http://www.w3.org/2000/svg\" width=\"24\" height=\"24\"><path d=\"m11.293 5.293 1.414 1.414-8 8-1.414-1.414 8-8Zm7-1 1.414 1.414-8 8-1.414-1.414 8-8Zm0 6 1.414 1.414-8 8-1.414-1.414 8-8Z\"></path></svg><div><div class=\"font-inter-tight text-lg font-semibold text-zinc-200 mb-1\">Scale and align your design team</div><div class=\"text-zinc-500\">Save time and keep things consistent with reusable images, and 3D assets in shared libraries.</div></div></button></div></div><!-- Tabs items --><div class=\"relative lg:max-w-none\"><div class=\"relative flex flex-col\"><!-- Item 1 --><div class=\"w-full\" x-show=\"tab === &#39;1&#39;\" x-transition:enter=\"transition ease-in-out duration-700 transform order-first\" x-transition:enter-start=\"opacity-0 translate-x-8\" x-transition:enter-end=\"opacity-100 translate-x-0\" x-transition:leave=\"transition ease-in-out duration-300 transform absolute\" x-transition:leave-start=\"opacity-100 translate-x-0\" x-transition:leave-end=\"opacity-0 -translate-x-8\"><div><img class=\"lg:max-w-none mx-auto rounded-lg shadow-2xl\" src=\"./images/carousel-illustration-01.jpg\" width=\"800\" height=\"620\" alt=\"Carousel 01\"></div></div><!-- Item 2 --><div class=\"w-full\" x-show=\"tab === &#39;2&#39;\" x-transition:enter=\"transition ease-in-out duration-700 transform order-first\" x-transition:enter-start=\"opacity-0 translate-x-8\" x-transition:enter-end=\"opacity-100 translate-x-0\" x-transition:leave=\"transition ease-in-out duration-300 transform absolute\" x-transition:leave-start=\"opacity-100 translate-x-0\" x-transition:leave-end=\"opacity-0 -translate-x-8\"><div><img class=\"lg:max-w-none mx-auto rounded-lg shadow-2xl\" src=\"./images/carousel-illustration-01.jpg\" width=\"800\" height=\"620\" alt=\"Carousel 02\"></div></div><!-- Item 3 --><div class=\"w-full\" x-show=\"tab === &#39;3&#39;\" x-transition:enter=\"transition ease-in-out duration-700 transform order-first\" x-transition:enter-start=\"opacity-0 translate-x-8\" x-transition:enter-end=\"opacity-100 translate-x-0\" x-transition:leave=\"transition ease-in-out duration-300 transform absolute\" x-transition:leave-start=\"opacity-100 translate-x-0\" x-transition:leave-end=\"opacity-0 -translate-x-8\"><div><img class=\"lg:max-w-none mx-auto rounded-lg shadow-2xl\" src=\"./images/carousel-illustration-01.jpg\" width=\"800\" height=\"620\" alt=\"Carousel 03\"></div></div></div><!-- Gear illustration --><img class=\"absolute left-0 bottom-0 -translate-x-1/2 translate-y-1/3 mix-blend-exclusion max-lg:w-32\" src=\"https://cdn.sonr.id/img/secure-vault.svg\" alt=\"Features 02 illustration\" width=\"224\" height=\"224\" aria-hidden=\"true\"></div></div></div>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = featuresBlocks().Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</div></section>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return templ_7745c5c3_Err
	})
}

func featuresBlocks() templ.Component {
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
		templ_7745c5c3_Var2 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var2 == nil {
			templ_7745c5c3_Var2 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<div class=\"max-w-6xl mx-auto px-4 sm:px-6 mt-24 lg:mt-32\"><div class=\"grid sm:grid-cols-2 lg:grid-cols-3 gap-8 lg:gap-16\"><!-- Block #1 --><div><div class=\"flex items-center mb-1\"><svg class=\"fill-zinc-400 mr-2\" xmlns=\"http://www.w3.org/2000/svg\" width=\"16\" height=\"16\"><path d=\"M15 9a1 1 0 0 1 0 2c-.441 0-1.243.92-1.89 1.716.319 1.005.529 1.284.89 1.284a1 1 0 0 1 0 2 2.524 2.524 0 0 1-2.339-1.545A3.841 3.841 0 0 1 9 16a1 1 0 0 1 0-2c.441 0 1.243-.92 1.89-1.716C10.57 11.279 10.361 11 10 11a1 1 0 0 1 0-2 2.524 2.524 0 0 1 2.339 1.545A3.841 3.841 0 0 1 15 9Zm-5-1H7.51l-.02.142C6.964 11.825 6.367 16 3 16a3 3 0 0 1-3-3 1 1 0 0 1 2 0 1 1 0 0 0 1 1c1.49 0 1.984-2.48 2.49-6H3a1 1 0 1 1 0-2h2.793c.52-3.1 1.4-6 4.207-6a3 3 0 0 1 3 3 1 1 0 0 1-2 0 1 1 0 0 0-1-1C8.808 2 8.257 3.579 7.825 6H10a1 1 0 0 1 0 2Z\"></path></svg><h3 class=\"font-inter-tight font-semibold text-zinc-200\">Multi-party Computation</h3></div><p class=\"text-sm text-zinc-500\">Eliminate the need for seed-phrases, and browser extensions for crypto wallets.</p></div><!-- Block #2 --><div><div class=\"flex items-center mb-1\"><svg class=\"fill-zinc-400 mr-2\" xmlns=\"http://www.w3.org/2000/svg\" width=\"16\" height=\"16\"><path d=\"M13 16c-.153 0-.306-.035-.447-.105l-3.851-1.926c-.231.02-.465.031-.702.031-4.411 0-8-3.14-8-7s3.589-7 8-7 8 3.14 8 7c0 1.723-.707 3.351-2 4.63V15a1.003 1.003 0 0 1-1 1Zm-4.108-4.054c.155 0 .308.036.447.105L12 13.382v-2.187c0-.288.125-.562.341-.752C13.411 9.506 14 8.284 14 7c0-2.757-2.691-5-6-5S2 4.243 2 7s2.691 5 6 5c.266 0 .526-.02.783-.048a1.01 1.01 0 0 1 .109-.006Z\"></path></svg><h3 class=\"font-inter-tight font-semibold text-zinc-200\">Matrix Protocol</h3></div><p class=\"text-sm text-zinc-500\">End-to-end encrypted messaging with <a class=\"text-zinc-500 hover:text-zinc-300 transition\" href=\"https://sonr.chat\">sonr.chat</a>.</p></div><!-- Block #3 --><div><div class=\"flex items-center mb-1\"><svg class=\"fill-zinc-400 mr-2\" xmlns=\"http://www.w3.org/2000/svg\" width=\"14\" height=\"16\"><path d=\"M13 0H1C.4 0 0 .4 0 1v14c0 .6.4 1 1 1h8l5-5V1c0-.6-.4-1-1-1ZM2 2h10v8H8v4H2V2Z\"></path></svg><h3 class=\"font-inter-tight font-semibold text-zinc-200\">OpenID Connect</h3></div><p class=\"text-sm text-zinc-500\">Seamless integration with Web2 Applications with the OpenID Standard.</p></div><!-- Block #4 --><div><div class=\"flex items-center mb-1\"><svg class=\"fill-zinc-400 mr-2\" xmlns=\"http://www.w3.org/2000/svg\" width=\"16\" height=\"16\"><path d=\"M7 14c-3.86 0-7-3.14-7-7s3.14-7 7-7 7 3.14 7 7-3.14 7-7 7ZM7 2C4.243 2 2 4.243 2 7s2.243 5 5 5 5-2.243 5-5-2.243-5-5-5Zm8.707 12.293a.999.999 0 1 1-1.414 1.414L11.9 13.314a8.019 8.019 0 0 0 1.414-1.414l2.393 2.393Z\"></path></svg><h3 class=\"font-inter-tight font-semibold text-zinc-200\">Native ETH & BTC</h3></div><p class=\"text-sm text-zinc-500\">Keep workflows efficient with tools that give teams visibility throughout the process.</p></div><!-- Block #5 --><div><div class=\"flex items-center mb-1\"><svg class=\"fill-zinc-400 mr-2\" xmlns=\"http://www.w3.org/2000/svg\" width=\"16\" height=\"16\"><path d=\"M14.6.085 8 2.885 1.4.085c-.5-.2-1.4-.1-1.4.9v11c0 .4.2.8.6.9l7 3c.3.1.5.1.8 0l7-3c.4-.2.6-.5.6-.9v-11c0-1-.9-1.1-1.4-.9ZM2 2.485l5 2.1v8.8l-5-2.1v-8.8Zm12 8.8-5 2.1v-8.7l5-2.1v8.7Z\"></path></svg><h3 class=\"font-inter-tight font-semibold text-zinc-200\">Developer Integrations</h3></div><p class=\"text-sm text-zinc-500\">Quickly integrate with Sonr by leveraging modern web APIs.</p></div><!-- Block #6 --><div><div class=\"flex items-center mb-1\"><svg class=\"fill-zinc-400 mr-2\" xmlns=\"http://www.w3.org/2000/svg\" width=\"14\" height=\"16\"><path d=\"M13 14a1 1 0 0 1 0 2H1a1 1 0 0 1 0-2h12Zm-6.707-2.293-5-5a1 1 0 0 1 1.414-1.414L6 8.586V1a1 1 0 1 1 2 0v7.586l3.293-3.293a1 1 0 1 1 1.414 1.414l-5 5a1 1 0 0 1-1.414 0Z\"></path></svg><h3 class=\"font-inter-tight font-semibold text-zinc-200\">Secure Backups</h3></div><p class=\"text-sm text-zinc-500\">Uniformly accessible user keyshare access on/off network.</p></div></div></div>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return templ_7745c5c3_Err
	})
}

var _ = templruntime.GeneratedTemplate