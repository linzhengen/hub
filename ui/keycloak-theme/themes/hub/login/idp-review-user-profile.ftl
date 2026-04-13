<#import "template.ftl" as layout>
<@layout.registrationLayout displayMessage=!messagesPerField.existsError('username','email','firstName','lastName'); section>
    <#if section = "header">
        ${msg("loginIdpReviewProfileTitle")}
    <#elseif section = "form">
        <form id="kc-idp-review-profile-form" class="space-y-4" action="${url.loginAction}" method="post">

            <#if !realm.registrationEmailAsUsername>
                <div>
                    <label for="username" class="sr-only">${msg("username")}</label>
                    <input type="text" id="username" name="username" value="${(user.username!'')}" autocomplete="username"
                           placeholder="${msg("username")}"
                           class="w-full bg-transparent border border-gray-500/30 dark:border-gray-400/50 outline-none rounded-full py-2.5 px-4 focus:ring-1 focus:ring-indigo-500 dark:focus:ring-indigo-400 shadow-sm dark:shadow-gray-900/30 focus:shadow-md transition-all"
                           <#if user.editUsernameAllowed>autofocus<#else>disabled</#if>
                           aria-invalid="<#if messagesPerField.existsError('username')>true</#if>"
                    />
                    <#if messagesPerField.existsError('username')>
                        <p class="mt-1 ml-4 text-xs text-rose-600 dark:text-rose-400" id="input-error-username" aria-live="polite">
                            ${kcSanitize(messagesPerField.getFirstError('username'))?no_esc}
                        </p>
                    </#if>
                </div>
            </#if>

            <div>
                <label for="email" class="sr-only">${msg("email")}</label>
                <input type="text" id="email" name="email" value="${(user.email!'')}" autocomplete="email"
                       placeholder="${msg("email")}"
                       class="w-full bg-transparent border border-gray-500/30 dark:border-gray-400/50 outline-none rounded-full py-2.5 px-4 focus:ring-1 focus:ring-indigo-500 dark:focus:ring-indigo-400 shadow-sm dark:shadow-gray-900/30 focus:shadow-md transition-all"
                       <#if !user.editEmailAllowed>disabled</#if>
                       aria-invalid="<#if messagesPerField.existsError('email')>true</#if>"
                />
                <#if messagesPerField.existsError('email')>
                    <p class="mt-1 ml-4 text-xs text-rose-600 dark:text-rose-400" id="input-error-email" aria-live="polite">
                        ${kcSanitize(messagesPerField.getFirstError('email'))?no_esc}
                    </p>
                </#if>
            </div>

            <div class="flex gap-4">
                <div class="w-1/2">
                    <label for="firstName" class="sr-only">${msg("firstName")}</label>
                    <input type="text" id="firstName" name="firstName" value="${(user.firstName!'')}"
                           placeholder="${msg("firstName")}"
                           class="w-full bg-transparent border border-gray-500/30 dark:border-gray-400/50 outline-none rounded-full py-2.5 px-4 focus:ring-1 focus:ring-indigo-500 dark:focus:ring-indigo-400 shadow-sm dark:shadow-gray-900/30 focus:shadow-md transition-all"
                           aria-invalid="<#if messagesPerField.existsError('firstName')>true</#if>"
                    />
                    <#if messagesPerField.existsError('firstName')>
                        <p class="mt-1 ml-4 text-xs text-rose-600 dark:text-rose-400" id="input-error-firstname" aria-live="polite">
                            ${kcSanitize(messagesPerField.getFirstError('firstName'))?no_esc}
                        </p>
                    </#if>
                </div>
                <div class="w-1/2">
                    <label for="lastName" class="sr-only">${msg("lastName")}</label>
                    <input type="text" id="lastName" name="lastName" value="${(user.lastName!'')}"
                           placeholder="${msg("lastName")}"
                           class="w-full bg-transparent border border-gray-500/30 dark:border-gray-400/50 outline-none rounded-full py-2.5 px-4 focus:ring-1 focus:ring-indigo-500 dark:focus:ring-indigo-400 shadow-sm dark:shadow-gray-900/30 focus:shadow-md transition-all"
                           aria-invalid="<#if messagesPerField.existsError('lastName')>true</#if>"
                    />
                    <#if messagesPerField.existsError('lastName')>
                        <p class="mt-1 ml-4 text-xs text-rose-600 dark:text-rose-400" id="input-error-lastname" aria-live="polite">
                            ${kcSanitize(messagesPerField.getFirstError('lastName'))?no_esc}
                        </p>
                    </#if>
                </div>
            </div>

            <div id="kc-form-buttons">
                <button class="w-full bg-indigo-500 dark:bg-indigo-600 py-2.5 rounded-full text-white font-medium hover:bg-indigo-600 dark:hover:bg-indigo-500 shadow-md hover:shadow-lg dark:shadow-indigo-900/50 transition-colors cursor-pointer" type="submit">
                    ${msg("doSubmit")}
                </button>
            </div>
        </form>
    </#if>
</@layout.registrationLayout>
