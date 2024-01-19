schema_version = 1

project {
  license        = "MIT"
  copyright_year = 2023
  copyright_holder = "Omlox Client Go Contributors"

  # (OPTIONAL) A list of globs that should not have copyright/license headers.
  # Supports doublestar glob patterns for more flexibility in defining which
  # files or folders should be ignored
  header_ignore = [
    "**.yaml",
    "**.yml",
    # "vendor/**",
    # "**autogen**",
  ]
}
