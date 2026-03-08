# Android Signing Keys

## Debug Keystore (debug.keystore)

- **Alias:** helixagent-debug
- **Store Password:** helixagent-debug
- **Key Password:** helixagent-debug
- **Algorithm:** RSA 2048-bit
- **Validity:** 10000 days
- **DN:** CN=HelixAgent Debug, O=HelixAgent, L=Dev, C=US

This keystore is for development/CI builds only. Apps signed with this key
can be installed on devices but are not suitable for Play Store distribution.

## Release Signing

Mount a release keystore at runtime via environment variables:

```bash
CI_ANDROID_RELEASE_KEYSTORE=/path/to/release.keystore \
CI_KEYSTORE_PASSWORD=secret \
CI_KEY_ALIAS=release \
CI_KEY_PASSWORD=secret \
make ci-mobile
```

The CI pipeline auto-selects: if `CI_ANDROID_RELEASE_KEYSTORE` is set, uses
release keystore; otherwise falls back to debug.keystore.
