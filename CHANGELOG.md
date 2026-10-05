## v0.15.0 (Mon, 05 Oct 2026 20:23:24 UTC)
- build(deps): update ctx42 dependencies to latest releases.
- fix(spec)!: return untyped nil from NewFieldError for nil error.
- style(spec): reindent over-wide JSON fixture in encode test.
- fix(spec): return ErrInvSpec when decoding into a nil spec.
- fix(spec): reset the target spec before decoding.
- fix(spec): reject specs with an empty name.
- fix(spec): make the zero value Registry usable.
- fix(spec): encode plain arguments with the registry's jsontype types.
- fix(spec): name the argument in plain argument encode errors.
- fix(spec): keep the decode error cause when wrapping sentinels.
- fix(spec): reject unnamed sources and encoding non-go sources.
- fix(spec): reject encoding a value matching several sources.
- fix(spec): keep empty lists and null arguments in JSON round trip.
- fix(spec): detect cyclic specs during encoding.
- test(spec): strengthen source lookups and cover registry guarantees.
- style(spec): tidy registry errors, locals and godoc.
- style(spec): align tests with the project test style.
- style(verax): fit lines, normalize nolint, space switch cases.
- fix(verax): skip nil elements when validating slices and maps.
- fix(verax): return an error from EnsureString for nil input.
- fix(verax): check map key assignability in the right direction.
- fix(verax): format non-integer map keys in Each errors.
- fix(verax): stop Contain matching empty elements.
- fix(verax): apply custom message and code in Contain errors.
- fix(verax): return the sticky error from Contain.
- fix(verax): let RangeRule.With replace the unsupported type error.
- fix(verax): keep a false SkipRule through a spec round trip.
- fix(verax): rebuild Set, Each and Key specs without rules.
- fix(verax): keep the cause of wrapped errors.
- test(verax): use a distinct custom code in error override tests.
- fix(verax): keep a custom RangeRule message in Exclusive.
- fix(verax): accept any sign magnitude in exclusive range checks.
- fix(verax): reject NaN in float range comparisons.
- fix(verax): stop using the Check message as a format string.
- fix(verax): accept named types and pointers in Check.
- fix(verax): build rules from constructors returning an interface.
- fix(verax): accept unnamed functions for named function spec arguments.
- fix(verax): report invalid Length bounds as an invalid range.
- fix(verax): check the MapRule condition first and allow nil map pointers.
- fix(verax): treat an untyped nil as an empty collection.
- test(verax): make ValidateWith and SkipRule.When tests able to fail.
- refactor(verax): remove the unused hasGoSource helper.
- fix(verax): skip nil rules in Validate.
- fix(verax): check the type of typed nil values in TypeRule.
- fix(verax): format values in custom messages like default messages.
- fix(verax): allow template delimiters in EqualField field names.
- fix(verax)!: return error from NewFieldError.
- refactor(verax): remove redundant code.
- refactor(verax): declare message templates as constants.
- style(verax): name receivers with three-letter abbreviations.
- test(verax): unexport test fixtures.
- test(verax): rename tests and subtests to the naming rules.
- test(verax): structure tests with Given, When and Then.
- test(verax): name expected values want and inline short literals.
- docs(verax): fix and complete godoc.
- style(rule): tidy declarations and godoc.
- style(rule): tidy tests.
- fix(rule): reject dns names with an all-numeric last label.
- fix(rule): limit dns names to 253 characters.
- fix(rule): stop dns name suffix from extending the last label.
- fix(rule): accept uppercase letters at the end of domain labels.
- fix(rule): limit domain names to 253 characters.
- fix(rule): accept uppercase and hyphenated punycode tlds.
- fix(rule): reject ipv4-mapped ipv6 addresses in IsIPv4.
- fix(rule): reject port numbers with a sign or leading zeros.
- refactor(rule): simplify the base64 check.
- docs: list the rules the rule package actually provides.
- style(verax): drop unused fixture and cmp shadow in tests.
- docs(readme): restructure and inject every example from code.

## v0.14.0 (Mon, 13 Jul 2026 21:12:12 UTC)
- chore: update AGENTS.md file.
- fix(rule): match punycode TLDs in domain validation.
- fix(verax): honor Contain().When(false) on empty collections.
- fix(spec): guard EncodeSpec against a nil Spec.
- fix(spec): report the real argument name on decode failure.
- docs(spec): add README and runnable example.
- refactor(verax): apply review cleanups.
- refactor(rule)!: rename IsSemver to IsSemVer.
- chore: update mirror and testing dependencies.
- test(spec): assert the missing signed-int converters.
- test(verax): drop unused sink var and reorder Set tests.
- test(rule): name domain and semver cases descriptively.

## v0.13.0 (Sun, 07 Jun 2026 16:11:48 UTC)
- perf: return findStructField by value to eliminate heap allocation.
- chore: Update dependencies.

## v0.12.0 (Mon, 01 Jun 2026 21:25:42 UTC)
- doc: update dependencies, add AGENTS.md with project guidelines and update .editorconfig.
- chore: improve documentation, publishing hygiene, and spec safety.

## v0.11.0 (Mon, 25 May 2026 08:13:06 UTC)
- perf: eliminate hot-path allocations in core validation.

## v0.10.0 (Sun, 24 May 2026 19:51:39 UTC)
- test: add JSON decode tests for all Specable rules.
- doc: Update Readme.md file.

## v0.9.0 (Tue, 19 May 2026 20:26:50 UTC)
- test: add JSON representation tests for Specable rules.
- test: use reg.EncodeSpec in JSON representation tests.

## v0.7.1 (Sat, 09 May 2026 19:42:20 UTC)
- test(error_types): add Test_NewInternalErrorf.
- test: improve coverage for registry and validation rules.

## v0.7.0 (Sat, 09 May 2026 13:14:23 UTC)
- refactor(error): use format-string error constructors from xrr upgrade.
- docs(readme): replace detailed guide with concise intro.

## v0.6.0 (Fri, 08 May 2026 15:17:25 UTC)
- feat: add Registry.Build, DecodeAndBuild, and Set serialization.

## v0.5.0 (Fri, 01 May 2026 21:27:38 UTC)
- fix!: Invalid merge.

## v0.4.0 (Fri, 01 May 2026 21:14:10 UTC)
- refactor(verax): overhaul error types and unify field error domain.
- refactor(verax): propagate FieldErrors rename across callsites.

## v0.3.0 (Sun, 26 Apr 2026 14:43:31 UTC)
- chore: Update dependencies.
- feat(spec): add spec package for serializable rule descriptions.
- feat(verax): overhaul error types, restructure rules, and implement Specable.
- docs(verax): fix incorrect error messages in README examples.
- chore: remove email from SPDX copyright headers.

## v0.2.0 (Fri, 17 Oct 2025 14:48:20 UTC)
- doc: Add missing SPDX file header.
- doc: Update README.md stryle.
- feat: Add `TypeRule` validating a value is of the expected type.

## v0.1.0 (Thu, 02 Oct 2025 11:04:46 UTC)
- Initial commit.
- doc: Update `README.md`.

