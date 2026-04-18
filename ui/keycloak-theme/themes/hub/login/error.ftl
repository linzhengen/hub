<#import "template.ftl" as layout>
<@layout.registrationLayout displayMessage=false; section>
    <#if section = "header">
        ${msg("errorTitle")}
    <#elseif section = "form">
        <div id="kc-error-message" class="text-center">
            <div class="mb-6 p-4 rounded-lg border bg-rose-50 border-rose-200 text-rose-800 dark:bg-rose-900/30 dark:border-rose-700 dark:text-rose-200 flex items-center gap-3">
                <div class="shrink-0">
                    <span class="text-xl">❌</span>
                </div>
                <span class="text-sm font-medium text-left">${message.summary?no_esc}</span>
            </div>

            <#if client?? && client.baseUrl?has_content>
                <a id="backToApplication" href="${client.baseUrl}" class="inline-block w-full bg-indigo-500 dark:bg-indigo-600 py-2.5 rounded-full text-white font-medium hover:bg-indigo-600 dark:hover:bg-indigo-500 shadow-md hover:shadow-lg dark:shadow-indigo-900/50 transition-colors cursor-pointer text-center">
                    ${msg("backToApplication")}
                </a>
            <#else>
                <a id="backToLogin" href="${url.loginUrl}" class="inline-block w-full bg-indigo-500 dark:bg-indigo-600 py-2.5 rounded-full text-white font-medium hover:bg-indigo-600 dark:hover:bg-indigo-500 shadow-md hover:shadow-lg dark:shadow-indigo-900/50 transition-colors cursor-pointer text-center">
                    ${msg("backToLogin")}
                </a>
            </#if>
        </div>
    </#if>
</@layout.registrationLayout>
