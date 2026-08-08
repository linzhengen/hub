<#import "template.ftl" as layout>
<@layout.registrationLayout; section>
    <#if section = "header">
        ${msg("oauth2DeviceVerificationTitle")}
    <#elseif section = "form">
        <form id="kc-user-verify-device-user-code-form" class="space-y-4" action="${url.oauth2DeviceVerificationAction}" method="post">
            <div>
                <label for="device-user-code" class="sr-only">${msg("verifyOAuth2DeviceUserCode")}</label>
                <input id="device-user-code" name="device_user_code" autocomplete="off" type="text" autofocus dir="ltr"
                       placeholder="${msg("verifyOAuth2DeviceUserCode")}"
                       class="w-full bg-transparent border border-gray-500/30 dark:border-gray-400/50 outline-none rounded-full py-2.5 px-4 focus:ring-1 focus:ring-indigo-500 dark:focus:ring-indigo-400 shadow-sm dark:shadow-gray-900/30 focus:shadow-md transition-all text-center tracking-widest text-lg font-mono"
                />
            </div>

            <div id="kc-form-buttons">
                <button class="w-full bg-indigo-500 dark:bg-indigo-600 py-2.5 rounded-full text-white font-medium hover:bg-indigo-600 dark:hover:bg-indigo-500 shadow-md hover:shadow-lg dark:shadow-indigo-900/50 transition-colors cursor-pointer" type="submit">
                    ${msg("doSubmit")}
                </button>
            </div>
        </form>
    </#if>
</@layout.registrationLayout>
