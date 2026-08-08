<#import "template.ftl" as layout>
<@layout.registrationLayout; section>
    <#if section = "header">
        ${msg("pageExpiredTitle")}
    <#elseif section = "form">
        <div class="space-y-3 text-sm text-gray-600 dark:text-gray-400 text-center">
            <p>
                ${msg("pageExpiredMsg1")}
                <a id="loginRestartLink" href="${url.loginRestartFlowUrl}"
                   class="text-indigo-600 hover:text-indigo-500 dark:text-indigo-400 dark:hover:text-indigo-300 underline">
                    ${msg("doClickHere")}
                </a>.
            </p>
            <p>
                ${msg("pageExpiredMsg2")}
                <a id="loginContinueLink" href="${url.loginAction}"
                   class="text-indigo-600 hover:text-indigo-500 dark:text-indigo-400 dark:hover:text-indigo-300 underline">
                    ${msg("doClickHere")}
                </a>.
            </p>
        </div>
    </#if>
</@layout.registrationLayout>
