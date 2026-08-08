<#import "template.ftl" as layout>
<@layout.registrationLayout; section>
    <#if section = "header">
        ${msg("logoutConfirmTitle")}
    <#elseif section = "form">
        <div class="space-y-4">
            <p class="text-sm text-gray-600 dark:text-gray-400 text-center">
                ${msg("logoutConfirmHeader")}
            </p>

            <form action="${url.logoutConfirmAction}" onsubmit="confirmLogout.disabled = true; return true;" method="POST">
                <input type="hidden" name="session_code" value="${logoutConfirm.code}">

                <button name="confirmLogout" id="kc-logout" type="submit"
                        class="w-full bg-rose-500 dark:bg-rose-600 py-2.5 rounded-full text-white font-medium hover:bg-rose-600 dark:hover:bg-rose-500 shadow-md hover:shadow-lg transition-colors cursor-pointer">
                    ${msg("doLogout")}
                </button>
            </form>

            <#if !logoutConfirm.skipLink>
                <#if (client.baseUrl)?has_content>
                    <p class="text-center">
                        <a href="${client.baseUrl}" class="text-sm text-indigo-600 hover:text-indigo-500 dark:text-indigo-400 dark:hover:text-indigo-300 underline">
                            ${msg("backToApplication")}
                        </a>
                    </p>
                </#if>
            </#if>
        </div>
    </#if>
</@layout.registrationLayout>
