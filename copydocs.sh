cp internal/services/*/*_data_source.md docs/data-sources/
for f in docs/data-sources/*_data_source.md; do mv "$f" "$(echo "$f" | sed s/_data_source//)"; done

cp internal/services/*/*_resource.md docs/resources/
for f in docs/resources/*_resource.md; do mv "$f" "$(echo "$f" | sed s/_resource//)"; done