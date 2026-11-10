package asset

// assetVersion is injected at build time from a hash of the published asset
// tree. The browser asset FS appends it to every request so cached files are
// invalidated whenever their contents change.
var assetVersion string
