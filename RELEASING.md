# Releasing

1. Add release entry to changelog
2. Open a PR with the above, and merge that into main
3. Create new annotated tag (`git tag -a`) on the merged commit with the new version (e.g. v2.2.3)
4. Push the tag upstream (this will kick off the release pipeline in CI)
5. Copy change log entry for newest version into draft GitHub release created as part of CI publish steps
