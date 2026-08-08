<#import "template.ftl" as layout>
<@layout.registrationLayout; section>
    <#if section = "header">
        <#if client.name?has_content>
            ${msg("oauthGrantTitle", advancedMsg(client.name))}
        <#else>
            ${msg("oauthGrantTitle", client.clientId)}
        </#if>
    <#elseif section = "form">
        <div class="space-y-4">
            <p class="text-sm font-medium text-gray-700 dark:text-gray-300">${msg("oauthGrantRequest")}</p>

            <ul class="space-y-2">
                <#if oauth.clientScopesRequested??>
                    <#list oauth.clientScopesRequested as clientScope>
                        <li class="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400">
                            <span class="text-indigo-500">✓</span>
                            <span>
                                <#if !clientScope.dynamicScopeParameter??>
                                    ${advancedMsg(clientScope.consentScreenText)}
                                <#else>
                                    ${advancedMsg(clientScope.consentScreenText)}: <strong>${clientScope.dynamicScopeParameter}</strong>
                                </#if>
                            </span>
                        </li>
                    </#list>
                </#if>
            </ul>

            <#if client.attributes.policyUri?? || client.attributes.tosUri??>
                <p class="text-xs text-gray-500 dark:text-gray-400">
                    <#if client.name?has_content>
                        ${msg("oauthGrantInformation", advancedMsg(client.name))}
                    <#else>
                        ${msg("oauthGrantInformation", client.clientId)}
                    </#if>
                    <#if client.attributes.tosUri??>
                        ${msg("oauthGrantReview")}
                        <a href="${client.attributes.tosUri}" target="_blank" class="text-indigo-600 hover:text-indigo-500 dark:text-indigo-400 underline">${msg("oauthGrantTos")}</a>
                    </#if>
                    <#if client.attributes.policyUri??>
                        ${msg("oauthGrantReview")}
                        <a href="${client.attributes.policyUri}" target="_blank" class="text-indigo-600 hover:text-indigo-500 dark:text-indigo-400 underline">${msg("oauthGrantPolicy")}</a>
                    </#if>
                </p>
            </#if>

            <form action="${url.oauthAction}" method="POST" class="flex gap-3">
                <input type="hidden" name="code" value="${oauth.code}">
                <button name="accept" id="kc-login" type="submit"
                        class="flex-1 bg-indigo-500 dark:bg-indigo-600 py-2.5 rounded-full text-white font-medium hover:bg-indigo-600 dark:hover:bg-indigo-500 shadow-md hover:shadow-lg dark:shadow-indigo-900/50 transition-colors cursor-pointer">
                    ${msg("doYes")}
                </button>
                <button name="cancel" id="kc-cancel" type="submit"
                        class="flex-1 border border-gray-500/30 dark:border-gray-400/50 py-2.5 rounded-full text-gray-700 dark:text-gray-300 font-medium hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors cursor-pointer">
                    ${msg("doNo")}
                </button>
            </form>
        </div>
    </#if>
</@layout.registrationLayout>
