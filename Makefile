msgfmt: lang/en/LC_MESSAGES/messages.mo lang/cy/LC_MESSAGES/messages.mo

lang/%/LC_MESSAGES/messages.mo: lang/%/LC_MESSAGES/messages.po
	msgfmt -o $@ $^

msgmerge: lang/en/LC_MESSAGES/messages.po lang/cy/LC_MESSAGES/messages.po

lang/%/LC_MESSAGES/messages.po: lang/messages.pot
	msgmerge $@ $^ -o $@

lang/messages.pot: $(wildcard example/pkg/*.go) $(wildcard example/template/*.gohtml)
	go run cmd/mopodofo-extract/main.go -v -pkg ./example/... -tmpl './example/**/*.gohtml'

# Remember to add
#
#   "Language: en\n"
#   "Content-Type: text/plain; charset=utf-8\n"
#   "Plural-Forms: <the rules>"
#
# to the .po files when regenerating.
.PHONY: clean
clean:
	rm -rf lang
