<#import "template.ftl" as layout>
<@layout.registrationLayout displayMessage=!messagesPerField.existsError('username','password') displayInfo=realm.password && realm.registrationAllowed && !registrationDisabled??; section>
    <#if section = "header">
        ${msg("loginAccountTitle")}
    <#elseif section = "form">
        <div id="kc-form">
          <div id="kc-form-wrapper">
            <#if realm.password>
                <form id="kc-form-login" class="space-y-4" onsubmit="login.disabled = true; return true;" action="${url.loginAction}" method="post">
                    <#if !usernameHidden??>
                        <div>
                            <label for="username" class="sr-only">
                                <#if !realm.loginWithEmailAllowed>${msg("username")}<#elseif !realm.registrationEmailAsUsername>${msg("usernameOrEmail")}<#else>${msg("email")}</#if>
                            </label>
                            <input tabindex="1" id="username" name="username" value="${(login.username!'')}" type="text" autofocus autocomplete="off"
                                   placeholder="<#if !realm.loginWithEmailAllowed>${msg("username")}<#elseif !realm.registrationEmailAsUsername>${msg("usernameOrEmail")}<#else>${msg("email")}</#if>"
                                   class="w-full bg-transparent border border-gray-500/30 dark:border-gray-400/50 outline-none rounded-full py-2.5 px-4 focus:ring-1 focus:ring-indigo-500 dark:focus:ring-indigo-400 shadow-sm dark:shadow-gray-900/30 focus:shadow-md transition-all"
                                   aria-invalid="<#if messagesPerField.existsError('username','password')>true</#if>"
                            />
                            <#if messagesPerField.existsError('username','password')>
                                <p class="mt-1 ml-4 text-xs text-rose-600 dark:text-rose-400" id="input-error" aria-live="polite">
                                    ${kcSanitize(messagesPerField.getFirstError('username','password'))?no_esc}
                                </p>
                            </#if>
                        </div>
                    </#if>

                    <div>
                        <div class="relative">
                            <label for="password" class="sr-only">${msg("password")}</label>
                            <input tabindex="2" id="password" name="password" type="password" autocomplete="off"
                                   placeholder="${msg("password")}"
                                   class="w-full bg-transparent border border-gray-500/30 dark:border-gray-400/50 outline-none rounded-full py-2.5 px-4 focus:ring-1 focus:ring-indigo-500 dark:focus:ring-indigo-400 shadow-sm dark:shadow-gray-900/30 focus:shadow-md transition-all"
                                   aria-invalid="<#if messagesPerField.existsError('username','password')>true</#if>"
                            />
                        </div>
                        <div class="text-right pt-2 pr-2">
                            <#if realm.resetPasswordAllowed>
                                <a tabindex="5" href="${url.loginResetCredentialsUrl}" class="text-blue-500 dark:text-blue-400 underline text-xs">
                                    ${msg("doForgotPassword")}
                                </a>
                            </#if>
                        </div>
                    </div>

                    <div class="flex items-center px-2">
                        <#if realm.rememberMe && !usernameHidden??>
                            <div class="flex items-center">
                                <input tabindex="3" id="rememberMe" name="rememberMe" type="checkbox" <#if login.rememberMe??>checked</#if>
                                       class="h-4 w-4 text-indigo-600 dark:text-indigo-400 focus:ring-indigo-500 dark:focus:ring-indigo-400 border-slate-300 dark:border-gray-500 rounded"
                                >
                                <label for="rememberMe" class="ml-2 block text-xs text-slate-700 dark:text-gray-300">${msg("rememberMe")}</label>
                            </div>
                        </#if>
                    </div>

                    <div id="kc-form-buttons">
                        <input type="hidden" id="id-token" name="id_token" value="${(login.idToken!'')}" />
                        <button tabindex="4" class="w-full bg-indigo-500 dark:bg-indigo-600 py-2.5 rounded-full text-white font-medium hover:bg-indigo-600 dark:hover:bg-indigo-500 shadow-md hover:shadow-lg dark:shadow-indigo-900/50 transition-colors cursor-pointer"
                                name="login" id="kc-login" type="submit">
                            ${msg("doLogIn")}
                        </button>
                    </div>
                </form>
            </#if>
            </div>
        </div>
    <#elseif section = "info" >
        <#if realm.password && realm.registrationAllowed && !registrationDisabled??>
            <p class="text-center text-gray-500 dark:text-gray-400">
                ${msg("noAccount")}
                <a tabindex="6" href="${url.registrationUrl}" class="text-blue-500 underline">
                    ${msg("doRegister")}
                </a>
            </p>
        </#if>
    <#elseif section = "socialProviders" >
        <#if realm.password && social?? && social.providers?has_content>
            <div id="kc-social-providers" class="mt-4 border-t border-gray-500/10 dark:border-gray-700 pt-4">
                <#--  <div class="relative mb-6">
                    <div class="absolute inset-0 flex items-center" aria-hidden="true">
                        <div class="w-full border-t border-slate-200"></div>
                    </div>
                    <div class="relative flex justify-center text-sm">
                        <span class="px-2 bg-white text-slate-500">${msg("identity-provider-login-label")}</span>
                    </div>
                </div>  -->

                <div class="space-y-3">
                    <#list social.providers as p>
                        <#if p.alias = "google">
                            <a data-once-link id="social-${p.alias}"
                               class="w-full flex items-center gap-2 justify-center bg-white border border-gray-500/30 dark:bg-gray-800 dark:border-gray-600 py-2.5 rounded-full text-gray-800 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
                               href="${p.loginUrl}">
                                <img class="h-4 w-4" src="https://raw.githubusercontent.com/prebuiltui/prebuiltui/main/assets/login/googleFavicon.png" alt="googleFavicon">
                                <span>${p.displayName!}</span>
                            </a>
                        <#elseif p.alias = "apple">
                            <a data-once-link id="social-${p.alias}"
                               class="w-full flex items-center gap-2 justify-center bg-black dark:bg-gray-900 py-2.5 rounded-full text-white hover:bg-gray-900 dark:hover:bg-gray-800 transition-colors"
                               href="${p.loginUrl}">
                                <img class="h-4 w-4" src="https://raw.githubusercontent.com/prebuiltui/prebuiltui/main/assets/login/appleLogo.png" alt="appleLogo">
                                <span>${p.displayName!}</span>
                            </a>
                        <#elseif p.alias = "github">
                            <a data-once-link id="social-${p.alias}"
                               class="w-full flex items-center gap-2 justify-center bg-[#24292e] dark:bg-[#1b1f23] py-2.5 rounded-full text-white hover:bg-[#1b1f23] dark:hover:bg-[#0d1117] transition-colors"
                               href="${p.loginUrl}">
                                <svg class="h-4 w-4 fill-current" viewBox="0 0 16 16" xmlns="http://www.w3.org/2000/svg">
                                    <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"></path>
                                </svg>
                                <span>${p.displayName!}</span>
                            </a>
                        <#else>
                            <a data-once-link id="social-${p.alias}"
                               class="w-full flex items-center gap-2 justify-center border border-gray-500/30 dark:border-gray-600 py-2.5 rounded-full text-gray-800 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors"
                               href="${p.loginUrl}">
                                <#if p.iconClasses?has_content>
                                    <i class="${properties.kcCommonLogoIdP!} ${p.iconClasses!}" aria-hidden="true"></i>
                                </#if>
                                <span>${p.displayName!}</span>
                            </a>
                        </#if>
                    </#list>
                </div>
            </div>
        </#if>
    </#if>
</@layout.registrationLayout>
