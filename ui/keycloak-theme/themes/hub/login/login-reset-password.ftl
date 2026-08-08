<#import "template.ftl" as layout>
<@layout.registrationLayout displayInfo=true displayMessage=!messagesPerField.existsError('username'); section>
    <#if section = "header">
        ${msg("emailForgotTitle")}
    <#elseif section = "form">
        <form id="kc-reset-password-form" class="space-y-4" action="${url.loginAction}" method="post">
            <div>
                <label for="username" class="sr-only">
                    <#if !realm.loginWithEmailAllowed>${msg("username")}<#elseif !realm.registrationEmailAsUsername>${msg("usernameOrEmail")}<#else>${msg("email")}</#if>
                </label>
                <input type="text" id="username" name="username" autofocus value="${(auth.attemptedUsername!'')}" dir="ltr"
                       placeholder="<#if !realm.loginWithEmailAllowed>${msg("username")}<#elseif !realm.registrationEmailAsUsername>${msg("usernameOrEmail")}<#else>${msg("email")}</#if>"
                       class="w-full bg-transparent border border-gray-500/30 dark:border-gray-400/50 outline-none rounded-full py-2.5 px-4 focus:ring-1 focus:ring-indigo-500 dark:focus:ring-indigo-400 shadow-sm dark:shadow-gray-900/30 focus:shadow-md transition-all"
                       aria-invalid="<#if messagesPerField.existsError('username')>true</#if>"
                />
                <#if messagesPerField.existsError('username')>
                    <p class="mt-1 ml-4 text-xs text-rose-600 dark:text-rose-400" id="input-error-username" aria-live="polite">
                        ${kcSanitize(messagesPerField.get('username'))?no_esc}
                    </p>
                </#if>
            </div>

            <div class="flex items-center justify-between px-1">
                <a href="${url.loginUrl}" class="text-sm text-indigo-600 hover:text-indigo-500 dark:text-indigo-400 dark:hover:text-indigo-300">
                    ${msg("backToLogin")}
                </a>
            </div>

            <div id="kc-form-buttons">
                <button class="w-full bg-indigo-500 dark:bg-indigo-600 py-2.5 rounded-full text-white font-medium hover:bg-indigo-600 dark:hover:bg-indigo-500 shadow-md hover:shadow-lg dark:shadow-indigo-900/50 transition-colors cursor-pointer" type="submit">
                    ${msg("doSubmit")}
                </button>
            </div>
        </form>
    <#elseif section = "info">
        <p class="text-center text-sm text-gray-500 dark:text-gray-400 mt-2">
            <#if realm.duplicateEmailsAllowed>
                ${msg("emailInstructionUsername")}
            <#else>
                ${msg("emailInstruction")}
            </#if>
        </p>
    </#if>
</@layout.registrationLayout>
