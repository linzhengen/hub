<#import "template.ftl" as layout>
<@layout.registrationLayout displayInfo=true; section>
    <#if section = "header">
        ${msg("emailVerifyTitle")}
    <#elseif section = "form">
        <p class="text-sm text-gray-600 dark:text-gray-400 text-center mb-4">
            <#if verifyEmail??>
                ${msg("emailVerifyInstruction1", verifyEmail)}
            <#else>
                ${msg("emailVerifyInstruction4", user.email)}
            </#if>
        </p>
        <#if isAppInitiatedAction??>
            <form id="kc-verify-email-form" class="space-y-3" action="${url.loginAction}" method="post">
                <div id="kc-form-buttons" class="flex flex-col gap-2">
                    <#if verifyEmail??>
                        <button class="w-full border border-gray-500/30 dark:border-gray-400/50 py-2.5 rounded-full text-gray-700 dark:text-gray-300 font-medium hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors cursor-pointer" type="submit">
                            ${msg("emailVerifyResend")}
                        </button>
                    <#else>
                        <button class="w-full bg-indigo-500 dark:bg-indigo-600 py-2.5 rounded-full text-white font-medium hover:bg-indigo-600 dark:hover:bg-indigo-500 shadow-md hover:shadow-lg dark:shadow-indigo-900/50 transition-colors cursor-pointer" type="submit">
                            ${msg("emailVerifySend")}
                        </button>
                    </#if>
                    <button class="w-full border border-gray-500/30 dark:border-gray-400/50 py-2.5 rounded-full text-gray-700 dark:text-gray-300 font-medium hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors cursor-pointer" type="submit" name="cancel-aia" value="true" formnovalidate>
                        ${msg("doCancel")}
                    </button>
                </div>
            </form>
        </#if>
    <#elseif section = "info">
        <#if !isAppInitiatedAction??>
            <p class="text-center text-sm text-gray-500 dark:text-gray-400">
                ${msg("emailVerifyInstruction2")}
                <br/>
                <a href="${url.loginAction}" class="text-indigo-600 hover:text-indigo-500 dark:text-indigo-400 dark:hover:text-indigo-300 underline">${msg("doClickHere")}</a> ${msg("emailVerifyInstruction3")}
            </p>
        </#if>
    </#if>
</@layout.registrationLayout>
