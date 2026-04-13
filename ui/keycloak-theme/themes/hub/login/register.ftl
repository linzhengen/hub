<#import "template.ftl" as layout>
<@layout.registrationLayout displayMessage=!messagesPerField.existsError('firstName','lastName','email','username','password','password-confirm'); section>
    <#if section = "header">
        ${msg("registerTitle")}
    <#elseif section = "form">
        <form id="kc-register-form" class="space-y-4" action="${url.registrationAction}" method="post">
            <#if !realm.registrationEmailAsUsername>
                <div>
                    <label for="username" class="sr-only">${msg("username")}</label>
                    <input type="text" id="username" name="username" value="${(register.formData.username!'')}" autocomplete="username"
                           placeholder="${msg("username")}"
                           class="w-full bg-transparent border border-gray-500/30 dark:border-gray-400/50 outline-none rounded-full py-2.5 px-4 focus:ring-1 focus:ring-indigo-500 dark:focus:ring-indigo-400 shadow-sm dark:shadow-gray-900/30 focus:shadow-md transition-all"
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
                <input type="text" id="email" name="email" value="${(register.formData.email!'')}" autocomplete="email"
                       placeholder="${msg("email")}"
                       class="w-full bg-transparent border border-gray-500/30 dark:border-gray-400/50 outline-none rounded-full py-2.5 px-4 focus:ring-1 focus:ring-indigo-500 dark:focus:ring-indigo-400 shadow-sm dark:shadow-gray-900/30 focus:shadow-md transition-all"
                       aria-invalid="<#if messagesPerField.existsError('email')>true</#if>"
                />
                <#if messagesPerField.existsError('email')>
                    <p class="mt-1 ml-4 text-xs text-rose-600 dark:text-rose-400" id="input-error-email" aria-live="polite">
                        ${kcSanitize(messagesPerField.getFirstError('email'))?no_esc}
                    </p>
                </#if>
            </div>

            <#if !realm.registrationEmailAsUsername>
                <div class="flex gap-4">
                    <div class="w-1/2">
                        <label for="firstName" class="sr-only">${msg("firstName")}</label>
                        <input type="text" id="firstName" name="firstName" value="${(register.formData.firstName!'')}"
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
                        <input type="text" id="lastName" name="lastName" value="${(register.formData.lastName!'')}"
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
            </#if>

            <#if passwordRequired??>
                <div>
                    <label for="password" class="sr-only">${msg("password")}</label>
                    <input type="password" id="password" name="password" autocomplete="new-password"
                           placeholder="${msg("password")}"
                           class="w-full bg-transparent border border-gray-500/30 dark:border-gray-400/50 outline-none rounded-full py-2.5 px-4 focus:ring-1 focus:ring-indigo-500 dark:focus:ring-indigo-400 shadow-sm dark:shadow-gray-900/30 focus:shadow-md transition-all"
                           aria-invalid="<#if messagesPerField.existsError('password','password-confirm')>true</#if>"
                    />
                    <#if messagesPerField.existsError('password')>
                        <p class="mt-1 ml-4 text-xs text-rose-600 dark:text-rose-400" id="input-error-password" aria-live="polite">
                            ${kcSanitize(messagesPerField.getFirstError('password'))?no_esc}
                        </p>
                    </#if>
                </div>

                <div>
                    <label for="password-confirm" class="sr-only">${msg("passwordConfirm")}</label>
                    <input type="password" id="password-confirm" name="password-confirm"
                           placeholder="${msg("passwordConfirm")}"
                           class="w-full bg-transparent border border-gray-500/30 dark:border-gray-400/50 outline-none rounded-full py-2.5 px-4 focus:ring-1 focus:ring-indigo-500 dark:focus:ring-indigo-400 shadow-sm dark:shadow-gray-900/30 focus:shadow-md transition-all"
                           aria-invalid="<#if messagesPerField.existsError('password-confirm')>true</#if>"
                    />
                    <#if messagesPerField.existsError('password-confirm')>
                        <p class="mt-1 ml-4 text-xs text-rose-600 dark:text-rose-400" id="input-error-password-confirm" aria-live="polite">
                            ${kcSanitize(messagesPerField.getFirstError('password-confirm'))?no_esc}
                        </p>
                    </#if>
                </div>
            </#if>

            <#if recaptchaRequired??>
                <div class="form-group">
                    <div class="${properties.kcInputWrapperClass!}">
                        <div class="g-recaptcha" data-size="compact" data-sitekey="${recaptchaSiteKey}"></div>
                    </div>
                </div>
            </#if>

            <div id="kc-form-buttons">
                <button class="w-full bg-indigo-500 dark:bg-indigo-600 py-2.5 rounded-full text-white font-medium hover:bg-indigo-600 dark:hover:bg-indigo-500 shadow-md hover:shadow-lg dark:shadow-indigo-900/50 transition-colors cursor-pointer" type="submit">
                    ${msg("doRegister")}
                </button>
            </div>
        </form>
    <#elseif section = "info" >
        <p class="text-center text-gray-500 dark:text-gray-400">
            ${msg("alreadyHaveAccount")}
            <a href="${url.loginUrl}" class="text-blue-500 underline">
                ${msg("doLogIn")}
            </a>
        </p>
    </#if>
</@layout.registrationLayout>
