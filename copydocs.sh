cp internal/services/*/*_data_source.md docs/data-sources/

for f in $(ls -1 docs/data-sources/*_data_source.md | sort -r); do mv "$f" "$(echo "$f" | sed s/_data_source//)"; done

if [ -f "docs/data-sources/external.md" ]; then
    rm docs/data-sources/external.md
fi

cp internal/services/*/*_resource.md docs/resources/
for f in docs/resources/*_resource.md; do mv "$f" "$(echo "$f" | sed s/_resource//)"; done