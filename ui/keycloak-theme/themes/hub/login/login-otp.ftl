<#import "template.ftl" as layout>
<@layout.registrationLayout displayMessage=!messagesPerField.existsError('totp'); section>
    <#if section = "header">
        ${msg("doLogIn")}
    <#elseif section = "form">
        <form id="kc-otp-login-form" class="space-y-4" onsubmit="login.disabled = true; return true;" action="${url.loginAction}" method="post">
            <#if otpLogin.userOtpCredentials?size gt 1>
                <div class="flex flex-col gap-2">
                    <#list otpLogin.userOtpCredentials as otpCredential>
                        <label class="flex items-center gap-3 p-3 border border-gray-500/30 dark:border-gray-400/50 rounded-xl cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors">
                            <input id="kc-otp-credential-${otpCredential?index}" type="radio" name="selectedCredentialId" value="${otpCredential.id}"
                                   class="h-4 w-4 text-indigo-600 dark:text-indigo-400 focus:ring-indigo-500 border-gray-300 dark:border-gray-500"
                                   <#if otpCredential.id == otpLogin.selectedCredentialId>checked</#if>
                            />
                            <span class="text-sm text-gray-700 dark:text-gray-300">${otpCredential.userLabel}</span>
                        </label>
                    </#list>
                </div>
            </#if>

            <div>
                <label for="otp" class="sr-only">${msg("loginOtpOneTime")}</label>
                <input id="otp" name="otp" autocomplete="one-time-code" type="text" autofocus dir="ltr"
                       placeholder="${msg("loginOtpOneTime")}"
                       class="w-full bg-transparent border border-gray-500/30 dark:border-gray-400/50 outline-none rounded-full py-2.5 px-4 focus:ring-1 focus:ring-indigo-500 dark:focus:ring-indigo-400 shadow-sm dark:shadow-gray-900/30 focus:shadow-md transition-all text-center tracking-widest font-mono"
                       aria-invalid="<#if messagesPerField.existsError('totp')>true</#if>"
                />
                <#if messagesPerField.existsError('totp')>
                    <p class="mt-1 ml-4 text-xs text-rose-600 dark:text-rose-400" id="input-error-otp-code" aria-live="polite">
                        ${kcSanitize(messagesPerField.get('totp'))?no_esc}
                    </p>
                </#if>
            </div>

            <div id="kc-form-buttons">
                <button name="login" id="kc-login" type="submit"
                        class="w-full bg-indigo-500 dark:bg-indigo-600 py-2.5 rounded-full text-white font-medium hover:bg-indigo-600 dark:hover:bg-indigo-500 shadow-md hover:shadow-lg dark:shadow-indigo-900/50 transition-colors cursor-pointer">
                    ${msg("doLogIn")}
                </button>
            </div>
        </form>
    </#if>
</@layout.registrationLayout>
