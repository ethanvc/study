#!/usr/bin/env python3
"""
Copy Go's encoding/json/v2 (and dependencies) into the project,
rewriting import paths and removing goexperiment build tags.

Usage:
    python3 copy_jsonv2.py --go-src /path/to/go/src [--output DIR]
"""

import argparse
import os
import re
import shutil
import sys

MODULE = "github.com/ethanvc/study/golangproj"
DEFAULT_OUTPUT = "golangproj/logjson/internal/gostd/json"

# Directories to copy (relative to encoding/json/).
# Each entry is (src_relative, dst_relative).
COPY_DIRS = [
    ("v2", "v2"),
    ("jsontext", "jsontext"),
    ("internal", "internal"),  # only top-level files
    ("internal/jsonflags", "internal/jsonflags"),
    ("internal/jsonopts", "internal/jsonopts"),
    ("internal/jsonwire", "internal/jsonwire"),
    ("internal/jsontest", "internal/jsontest"),
]

# Import path replacements (longest first to avoid partial matches).
IMPORT_REPLACEMENTS = [
    ("encoding/json/internal/jsonflags", f"{MODULE}/logjson/internal/gostd/json/internal/jsonflags"),
    ("encoding/json/internal/jsonopts", f"{MODULE}/logjson/internal/gostd/json/internal/jsonopts"),
    ("encoding/json/internal/jsonwire", f"{MODULE}/logjson/internal/gostd/json/internal/jsonwire"),
    ("encoding/json/internal/jsontest", f"{MODULE}/logjson/internal/gostd/json/internal/jsontest"),
    ("encoding/json/internal", f"{MODULE}/logjson/internal/gostd/json/internal"),
    ("encoding/json/jsontext", f"{MODULE}/logjson/internal/gostd/json/jsontext"),
    ("encoding/json/v2", f"{MODULE}/logjson/internal/gostd/json/v2"),
]

# stdlib-internal import that must be replaced with a third-party equivalent.
ZSTD_OLD_IMPORT = '"internal/zstd"'
ZSTD_NEW_IMPORT = '"github.com/klauspost/compress/zstd"'

BUILD_TAG_LINE = "//go:build goexperiment.jsonv2"


def rewrite_imports(content: str) -> str:
    """Replace encoding/json/* import paths with the forked module paths."""
    for old, new in IMPORT_REPLACEMENTS:
        content = content.replace(f'"{old}"', f'"{new}"')
    return content


def remove_build_tag(content: str) -> str:
    """Remove the //go:build goexperiment.jsonv2 line and the blank line after it."""
    lines = content.split("\n")
    out = []
    skip_next_blank = False
    for line in lines:
        if line.strip() == BUILD_TAG_LINE:
            skip_next_blank = True
            continue
        if skip_next_blank and line.strip() == "":
            skip_next_blank = False
            continue
        skip_next_blank = False
        out.append(line)
    return "\n".join(out)


def fix_zstd_usage(content: str) -> str:
    """
    Replace internal/zstd import and adjust the call site in testdata.go.
    internal/zstd.NewReader returns io.Reader directly, but
    klauspost/compress/zstd.NewReader returns (*Decoder, error).
    """
    content = content.replace(ZSTD_OLD_IMPORT, ZSTD_NEW_IMPORT)

    # Replace the call site pattern:
    #   zr := zstd.NewReader(bytes.NewReader(b))
    #   return mustGet(io.ReadAll(zr))
    # with:
    #   zr, err := zstd.NewReader(bytes.NewReader(b))
    #   if err != nil { panic(err) }
    #   defer zr.Close()
    #   return mustGet(io.ReadAll(zr))
    old_call = (
        "zr := zstd.NewReader(bytes.NewReader(b))\n"
        "\t\t\treturn mustGet(io.ReadAll(zr))"
    )
    new_call = (
        "zr, err := zstd.NewReader(bytes.NewReader(b))\n"
        "\t\t\tif err != nil { panic(err) }\n"
        "\t\t\tdefer zr.Close()\n"
        "\t\t\treturn mustGet(io.ReadAll(zr))"
    )
    content = content.replace(old_call, new_call)
    return content


def process_go_file(src_path: str, dst_path: str, is_jsontest_testdata: bool):
    """Read a .go file, apply transformations, write to dst."""
    with open(src_path, "r", encoding="utf-8") as f:
        content = f.read()

    content = remove_build_tag(content)
    content = rewrite_imports(content)

    if is_jsontest_testdata:
        content = fix_zstd_usage(content)

    os.makedirs(os.path.dirname(dst_path), exist_ok=True)
    with open(dst_path, "w", encoding="utf-8") as f:
        f.write(content)


def copy_binary_file(src_path: str, dst_path: str):
    """Copy a non-.go file as-is."""
    os.makedirs(os.path.dirname(dst_path), exist_ok=True)
    shutil.copy2(src_path, dst_path)


def copy_directory(src_dir: str, dst_dir: str, recurse_subdirs: bool, jsontest_dir: str):
    """
    Copy .go files (with transformations) and non-.go files from src_dir to dst_dir.
    If recurse_subdirs is False, only copy files (not subdirectories) from src_dir,
    except for 'testdata' which is always copied recursively.
    """
    copied = 0
    for entry in sorted(os.listdir(src_dir)):
        src_path = os.path.join(src_dir, entry)

        if os.path.isdir(src_path):
            if entry == "testdata":
                # Always recursively copy testdata (binary assets).
                dst_td = os.path.join(dst_dir, entry)
                shutil.copytree(src_path, dst_td, dirs_exist_ok=True)
                print(f"  copied testdata/ -> {dst_td}")
                copied += 1
            elif not recurse_subdirs:
                continue
            else:
                copied += copy_directory(
                    src_path, os.path.join(dst_dir, entry),
                    recurse_subdirs=True, jsontest_dir=jsontest_dir
                )
            continue

        if not os.path.isfile(src_path):
            continue

        dst_path = os.path.join(dst_dir, entry)

        if entry.endswith(".go"):
            is_jsontest_testdata = (
                os.path.normpath(src_dir) == os.path.normpath(jsontest_dir)
                and entry == "testdata.go"
            )
            process_go_file(src_path, dst_path, is_jsontest_testdata)
            copied += 1
        else:
            copy_binary_file(src_path, dst_path)
            copied += 1

    return copied


def main():
    parser = argparse.ArgumentParser(
        description="Copy Go's encoding/json/v2 into the project with import rewriting."
    )
    parser.add_argument(
        "--go-src", required=True,
        help="Path to Go source root (e.g. /path/to/go1.26.1/src)"
    )
    parser.add_argument(
        "--output", default=None,
        help=f"Output directory (default: {DEFAULT_OUTPUT} relative to workspace root)"
    )
    args = parser.parse_args()

    go_src = os.path.abspath(args.go_src)
    json_src = os.path.join(go_src, "encoding", "json")

    if not os.path.isdir(json_src):
        print(f"ERROR: {json_src} does not exist", file=sys.stderr)
        sys.exit(1)

    if args.output:
        output_base = os.path.abspath(args.output)
    else:
        script_dir = os.path.dirname(os.path.abspath(__file__))
        output_base = os.path.join(script_dir, "json")

    jsontest_dir = os.path.join(json_src, "internal", "jsontest")

    if os.path.exists(output_base):
        print(f"Cleaning existing output directory: {output_base}")
        shutil.rmtree(output_base)

    print(f"Source: {json_src}")
    print(f"Output: {output_base}")
    print()

    total = 0
    for src_rel, dst_rel in COPY_DIRS:
        src_dir = os.path.join(json_src, src_rel)
        dst_dir = os.path.join(output_base, dst_rel)

        if not os.path.isdir(src_dir):
            print(f"WARNING: source directory {src_dir} not found, skipping", file=sys.stderr)
            continue

        # For "internal" (top-level), don't recurse into subdirectories
        # because subdirs are listed separately in COPY_DIRS.
        recurse = src_rel != "internal"

        print(f"Copying {src_rel}/ -> {dst_rel}/")
        n = copy_directory(src_dir, dst_dir, recurse_subdirs=recurse, jsontest_dir=jsontest_dir)
        total += n

    print(f"\nDone. Copied/processed {total} files.")
    print("\nNext steps:")
    print("  cd golangproj && go mod tidy")
    print("  cd golangproj && go build ./logjson/...")
    print("  cd golangproj && go test ./logjson/internal/gostd/json/...")


if __name__ == "__main__":
    main()
