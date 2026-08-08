<#import "template.ftl" as layout>
<@layout.registrationLayout displayMessage=!messagesPerField.existsError('password','password-confirm'); section>
    <#if section = "header">
        ${msg("updatePasswordTitle")}
    <#elseif section = "form">
        <form id="kc-passwd-update-form" class="space-y-4" onsubmit="login.disabled = true; return true;" action="${url.loginAction}" method="post">
            <div>
                <label for="password-new" class="sr-only">${msg("passwordNew")}</label>
                <input type="password" id="password-new" name="password-new" autofocus autocomplete="new-password"
                       placeholder="${msg("passwordNew")}"
                       class="w-full bg-transparent border border-gray-500/30 dark:border-gray-400/50 outline-none rounded-full py-2.5 px-4 focus:ring-1 focus:ring-indigo-500 dark:focus:ring-indigo-400 shadow-sm dark:shadow-gray-900/30 focus:shadow-md transition-all"
                       aria-invalid="<#if messagesPerField.existsError('password','password-confirm')>true</#if>"
                />
                <#if messagesPerField.existsError('password')>
                    <p class="mt-1 ml-4 text-xs text-rose-600 dark:text-rose-400" id="input-error-password" aria-live="polite">
                        ${kcSanitize(messagesPerField.get('password'))?no_esc}
                    </p>
                </#if>
            </div>

            <div>
                <label for="password-confirm" class="sr-only">${msg("passwordConfirm")}</label>
                <input type="password" id="password-confirm" name="password-confirm" autocomplete="new-password"
                       placeholder="${msg("passwordConfirm")}"
                       class="w-full bg-transparent border border-gray-500/30 dark:border-gray-400/50 outline-none rounded-full py-2.5 px-4 focus:ring-1 focus:ring-indigo-500 dark:focus:ring-indigo-400 shadow-sm dark:shadow-gray-900/30 focus:shadow-md transition-all"
                       aria-invalid="<#if messagesPerField.existsError('password-confirm')>true</#if>"
                />
                <#if messagesPerField.existsError('password-confirm')>
                    <p class="mt-1 ml-4 text-xs text-rose-600 dark:text-rose-400" id="input-error-password-confirm" aria-live="polite">
                        ${kcSanitize(messagesPerField.get('password-confirm'))?no_esc}
                    </p>
                </#if>
            </div>

            <div id="kc-form-buttons">
                <#if isAppInitiatedAction??>
                    <button name="login" class="w-full bg-indigo-500 dark:bg-indigo-600 py-2.5 rounded-full text-white font-medium hover:bg-indigo-600 dark:hover:bg-indigo-500 shadow-md hover:shadow-lg dark:shadow-indigo-900/50 transition-colors cursor-pointer" type="submit">
                        ${msg("doSubmit")}
                    </button>
                    <button class="w-full mt-2 border border-gray-500/30 dark:border-gray-400/50 py-2.5 rounded-full text-gray-700 dark:text-gray-300 font-medium hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors cursor-pointer" type="submit" name="cancel-aia" value="true">
                        ${msg("doCancel")}
                    </button>
                <#else>
                    <button name="login" class="w-full bg-indigo-500 dark:bg-indigo-600 py-2.5 rounded-full text-white font-medium hover:bg-indigo-600 dark:hover:bg-indigo-500 shadow-md hover:shadow-lg dark:shadow-indigo-900/50 transition-colors cursor-pointer" type="submit">
                        ${msg("doSubmit")}
                    </button>
                </#if>
            </div>
        </form>
    </#if>
</@layout.registrationLayout>
