# Image enhancements

The original PNG/JPEG `image.info`, `image.convert` and `image.resize` keep their core defaults. Set `--engine imagemagick` to use the managed runtime for additional raster formats and EXIF-aware conversion/resizing. All other Operations on this page require ImageMagick.

```sh
fnsh pkg add imagemagick
fnsh image convert photo.png --engine imagemagick -o photo.webp
fnsh image resize photo.jpg --engine imagemagick --width 800 -o preview.jpg
fnsh image crop screenshot.png --x 20 --y 30 --width 400 --height 200 -o crop.png
fnsh image contact-sheet first.png --files second.png --columns 2 -o grid.png
fnsh image annotate photo.png --font local-font.ttf --text "FinishBit" -o labeled.png
```

The registry pins ImageMagick 7.1.2-31 portable Q16 packages for Windows x64/arm64. Windows x64 is verified by the real acceptance suite. macOS/Linux packages are not registered yet; installation explicitly reports unsupported platforms. Runtime archives retain their configuration and license files.

## Catalog

| Area | Operations |
| --- | --- |
| Runtime discovery | `image.formats` |
| Geometry | `image.crop`, `image.rotate`, `image.flip`, `image.orient`, `image.canvas` |
| Encoding | `image.compress`, `image.strip-metadata`, `image.icon` |
| Composition | `image.watermark`, `image.annotate`, `image.join`, `image.contact-sheet` |
| Pixels and colors | `image.flatten`, `image.transparent`, `image.grayscale`, `image.blur`, `image.sharpen`, `image.adjust`, `image.difference` |
| Animation | `image.gif-create`, `image.gif-split` |

## Behavior

- Allowed raster extensions are PNG, JPG/JPEG, WebP, GIF, TIF/TIFF, BMP, ICO, AVIF and HEIC/HEIF. Actual read/write support depends on the pinned build; `image.formats` reports available coders. Unsupported codecs fail with a structured provider error. SVG, PDF, remote URLs and pseudo-image generators are outside these input contracts.
- Input files are copied into a private working directory with generated names and explicit raster coder prefixes, so filename syntax does not become an ImageMagick expression. Processing does not rename or edit input files. Outputs are staged; replacement requires `--overwrite`. GIF extraction requires a new directory.
- Existing image APIs keep `engine=core` by default and do not automatically download dependencies. Their optional engine choice is described in the option contract rather than an unconditional package requirement. An unavailable ImageMagick runtime returns `dependency_missing` with an installation suggestion.
- Ordinary transformations process the first frame/page of a multi-frame raster. `image.info --engine imagemagick` reports the frame count. GIF creation accepts ordered still images; GIF splitting coalesces frames into complete PNG canvases. ICO generation creates 256/128/64/48/32/16 pixel variants.
- Input bounds: at most 100 files, 64 MiB per file, 512 MiB combined, 500 frames per input, 100 million total pixels per input, 32768 pixels per dimension. Canvas/contact-sheet creation has a 25-million-pixel bound. ImageMagick also receives memory/map/disk/time/thread/list limits. Jobs remain cancellable through the shared application context.
- Crops must lie completely inside the input. `image.canvas` centers the image and can pad or crop. Image joining preserves each input's size; a contact sheet fits thumbnails into the requested tile layout. EXIF correction is explicit via `image.orient`, or automatic in the ImageMagick branches of convert/resize.
- JPEG output flattens transparency onto white. `image.flatten` accepts an opaque hex background; `image.transparent` selects a color and tolerance. Difference images require equal dimensions and provide visual channel differences, not a perceptual similarity score.
- Annotation takes a local font file, literal text, size, position and hex color. This makes font selection explicit across platforms. Metadata stripping re-encodes pixels and removes profiles/comments; it is not a guarantee that image content contains no private information. Quality compression does not promise a target byte size or a smaller result for every image.

## Real acceptance tests

Set `FINISHBIT_TEST_HOME` to the managed runtime root and `FINISHBIT_TEST_FONT` to a local TTF/OTF font. On Windows the test defaults to `%WINDIR%/Fonts/arial.ttf`.

```sh
go test ./internal/raster -run TestImagesWithManagedTools -count=1 -v
```

Tests invoke the shared application service for every new Operation and each enhanced existing API. They check crop pixels, transparency, grayscale channels, identical-image differences, grid dimensions, GIF frame counts, icon sizes, EXIF orientation, WebP/TIFF/BMP/AVIF round trips, Unicode/special-character paths, cancellation and existing-output preservation.
