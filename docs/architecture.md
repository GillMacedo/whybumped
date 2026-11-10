# Architecture

    parse flags (--old DIR | --base REF, --new DIR, --expect, --format, --fail-on)
      --base REF: git rev-parse --verify REF^{commit}, git archive REF -- **/go.mod **/go.sum
                  untar into a temp dir (no links, no paths outside it); old side = same subdirectory
      --> for each side (loadSide):
            checkGoMod: go.mod exists, isn't empty, is UTF-8, ends with a newline,
                        has a module line and a go directive          --> else exit 2
            go mod graph (GOWORK=off)                                  --> exit != 0 --> exit 2
              parseGraph: every line is two fields, path@version or the main module,
                          go@ and toolchain@ nodes checked and dropped, one main module,
                          ends with a newline                          --> else exit 2
            go list -m -mod=readonly -json all (GOWORK=off)           --> exit != 0 --> exit 2
              parseModules: valid JSON stream, no Error, a version for every dependency
            crossCheck: max version per path over the nodes reachable in the graph
                        must equal go list's selection, path for path  --> else exit 2
      --> same main module on both sides, and dependencies on both sides --> else exit 2
      --> compare:
            go directive: 1.N to 1.N.0 --> go-respelled; raised --> go-raised, explained by the selected modules whose
                          go line is newer than the old one; lowered --> go-lowered
            toolchain line changed --> toolchain-changed
            per module path in either build list:
              same version   --> unchanged
              raised, added  --> explainUp: requirers of path@new other than the main module
                                   none --> requested (INFO), with the highest version any
                                            dependency needs
                                   some --> pulled (WARN), chainFor back to a module
                                            only go.mod requires
              lowered        --> explainDown: who required the old version, and where they went
              removed        --> explainRemoved
            requirementLines: go.mod lines added or dropped while the selection stayed
                              the same --> requirement-added (INFO), requirement-dropped (WARN)
            --expect: matching changes become INFO; other requested changes become WARN;
                      expected modules that didn't change --> expect-missed
            roots: count of pulled changes per chain end
      --> text or JSON --> exit code

- **The go command decides the versions.** whybumped never resolves versions itself. It reads the graph the go command printed and the build list it selected, and refuses to report if its own minimal version selection over the graph disagrees with `go list`. That check also catches a graph cut at a line break, because the missing edges change the selection.
- **The main module's edges don't explain anything.** Since Go 1.17, a tidy go.mod lists every module in the build, so the main module requires everything at its selected version. A change is explained by the other modules that require that exact version. If there are none, only go.mod asks for it, which is what an intended upgrade or `go get -u` looks like.
- **Unselected versions count.** In minimal version selection, the requirements of every version in the graph count, even a version that lost to a newer one. Chains prefer selected versions, but go through an unselected one when that's the only requirer, and the detail says so.
- **Read-only.** `go mod graph` takes no `-mod` flag but doesn't write `go.sum`; `go list -m` runs with `-mod=readonly`, so a missing `go.sum` entry is an error (exit 2), never a write.

Key choices are in `decisions/`.
