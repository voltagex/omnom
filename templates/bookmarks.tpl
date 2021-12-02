{{ define "content" }}
<div class="content">
    {{ if not .Bookmarks }}
    <h3 class="title">No public bookmarks found</h3>
    {{ else }}
    <h3 class="title">Public bookmarks ({{ .BookmarkCount }})</h3>
    {{ end }}
    <div class="content"><form action="" method="get">
        <details{{ if .HasSearch }} open{{ end }}>
            <summary>
                Search
            </summary>
            {{ block "textFilter" .}}{{ end }}
            {{ block "domainFilter" .}}{{ end }}
            {{ block "ownerFilter" .}}{{ end }}
            {{ block "tagFilter" .}}{{ end }}
            {{ block "dateFilter" .}}{{ end }}
            {{ block "submit" . }}{{ end }}
        </details>
    </form></div>
    {{ range .Bookmarks }}
        {{ block "public-bookmark" .}}{{ end }}
    {{ end }}
{{ block "paging" .}}{{ end }}
</div>
{{ end }}
