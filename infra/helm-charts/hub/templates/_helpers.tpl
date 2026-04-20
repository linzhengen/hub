{{/*
Copyright Linzhengen. All Rights Reserved.

*/}}

{{/*
Return the proper Hub image name
*/}}
{{- define "hub.image" -}}
{{ include "common.images.image" (dict "imageRoot" .Values.image "global" .Values.global) }}
{{- end -}}

{{/*
Return the proper Docker Image Registry Secret Names
*/}}
{{- define "hub.imagePullSecrets" -}}
{{- include "common.images.renderPullSecrets" (dict "images" (list .Values.image) "context" .) -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "hub.fullname" -}}
{{ include "common.names.fullname" . }}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "hub.chart" -}}
{{ include "common.names.chart" . }}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "hub.labels" -}}
{{ include "common.labels.standard" (dict "customLabels" .Values.commonLabels "context" .) }}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "hub.selectorLabels" -}}
{{ include "common.labels.matchLabels" (dict "customLabels" .Values.commonLabels "context" .) }}
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "hub.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
    {{ default (include "common.names.fullname" .) .Values.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccount.name }}
{{- end -}}
{{- end -}}

{{/*
Return the Hub configuration ConfigMap name.
*/}}
{{- define "hub.configmapName" -}}
{{- if .Values.existingConfigmap -}}
    {{- tpl .Values.existingConfigmap . -}}
{{- else -}}
    {{- printf "%s-configuration" (include "common.names.fullname" .) -}}
{{- end -}}
{{- end -}}

{{/*
Return the secret containing environment variables
*/}}
{{- define "hub.secretName" -}}
{{- if .Values.existingSecret -}}
    {{- tpl .Values.existingSecret . -}}
{{- else -}}
    {{- include "common.names.fullname" . -}}
{{- end -}}
{{- end -}}

{{/*
Compile all warnings into a single message.
*/}}
{{- define "hub.validateValues" -}}
{{- $messages := list -}}
{{- $messages = append $messages (include "hub.validateValues.database" .) -}}
{{- $messages = append $messages (include "hub.validateValues.secret" .) -}}
{{- $messages = without $messages "" -}}
{{- $message := join "\n" $messages -}}

{{- if $message -}}
{{-   printf "\nVALUES VALIDATION:\n%s" $message | fail -}}
{{- end -}}
{{- end -}}

{{/* Validate values of Hub - database */}}
{{- define "hub.validateValues.database" -}}
{{- if and (not .Values.database.host) (not .Values.database.existingSecret) -}}
hub: database
    You must specify either a database host (--set database.host=FOO)
    or an existing secret containing database credentials (--set database.existingSecret=BAR).
{{- end -}}
{{- end -}}

{{/* Validate values of Hub - secret */}}
{{- define "hub.validateValues.secret" -}}
{{- /* No validation needed - chart can create its own secret */ -}}
{{- end -}}
