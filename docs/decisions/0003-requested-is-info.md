# 0003: Changes only go.mod asks for are INFO unless --expect is given

**Decision:** a version that only the main module's go.mod requires is a `requested` change and INFO. A version that another module requires is `pulled` and WARN. With `--expect`, every requested change that isn't listed becomes WARN too.

**Why:** a requested change is usually the change the author meant to make (the Dependabot bump itself, or a direct `go get`). The pulled ones are the side effects a reviewer can't see in the pull request title. When the intended change is known, anything else is a surprise, so `--expect` makes the check strict.

**Alternatives:** treating every change as WARN made every dependency pull request fail and hid the side effects among the intended changes.
