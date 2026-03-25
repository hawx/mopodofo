example-msgfmt: example/lang/en/LC_MESSAGES/messages.mo example/lang/cy/LC_MESSAGES/messages.mo

example/lang/%/LC_MESSAGES/messages.mo: example/lang/%/LC_MESSAGES/messages.po
	msgfmt -o $@ $^

example-msgmerge: example/lang/en/LC_MESSAGES/messages.po example/lang/cy/LC_MESSAGES/messages.po

example/lang/%/LC_MESSAGES/messages.po: example/lang/messages.pot
	msgmerge $@ $^ -o $@

example/lang/messages.pot: $(wildcard example/**)
	go build ./cmd/mopodofo-extract
	./mopodofo-extract -v -pkg ./example/... -tmpl './example/**/*.gohtml' -out example/lang/messages.pot

# Remember to add
#
#   "Language: en\n"
#   "Content-Type: text/plain; charset=utf-8\n"
#   "Plural-Forms: <the rules>"
#
# to the .po files when regenerating.
.PHONY: clean
clean:
	rm -rf example/lang
