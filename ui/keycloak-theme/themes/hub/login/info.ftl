<#import "template.ftl" as layout>
<@layout.registrationLayout displayMessage=false; section>
    <#if section = "header">
        <#if messageHeader??>
            ${kcSanitize(msg("${messageHeader}"))?no_esc}
        <#else>
            ${message.summary}
        </#if>
    <#elseif section = "form">
        <div id="kc-info-message" class="space-y-4 text-center">
            <p class="text-sm text-gray-600 dark:text-gray-400">
                ${message.summary}<#if requiredActions??><#list requiredActions>: <strong><#items as reqActionItem>${kcSanitize(msg("requiredAction.${reqActionItem}"))?no_esc}<#sep>, </#items></strong></#list></#if>
            </p>
            <#if skipLink??>
            <#else>
                <#if pageRedirectUri?has_content>
                    <a href="${pageRedirectUri}" class="text-indigo-600 hover:text-indigo-500 dark:text-indigo-400 dark:hover:text-indigo-300 underline text-sm">
                        ${msg("backToApplication")}
                    </a>
                <#elseif actionUri?has_content>
                    <a href="${actionUri}" class="text-indigo-600 hover:text-indigo-500 dark:text-indigo-400 dark:hover:text-indigo-300 underline text-sm">
                        ${msg("proceedWithAction")}
                    </a>
                <#elseif (client.baseUrl)?has_content>
                    <a href="${client.baseUrl}" class="text-indigo-600 hover:text-indigo-500 dark:text-indigo-400 dark:hover:text-indigo-300 underline text-sm">
                        ${msg("backToApplication")}
                    </a>
                </#if>
            </#if>
        </div>
    </#if>
</@layout.registrationLayout>
