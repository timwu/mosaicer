# Mosaicer

This is a simple CLI tool for creating a photo mosaic out of a collection of source images.

## How to use Mosaicer

1. (Only need to do this once) Index your given collection of source images:

   ```shell
   mosaicer index path/to/collection
   ```

1. Build a photo mosaic for a target image:

   ```shell
   mosaicer build --source path/to/collection target_image.jpg
   ```

This will produce the image `target_image.jpg.mosaic.jpg` with the best matching source images as tiles.

## How does this work?

`mosaicer` works in 2 phases: indexing and building. 

The indexing phase creates a data blob with downsampled versions of the source images in both RGB and L*A*B* colorspaces. It'll always have a 1x1 pixel version, then depending on the aspect ratio will have multiples of the aspect ratio. For example a given source image of aspect ratio 4:3, `--samples 3`, will have: 1x1, 4x3, 8x6 downsample images in the data blob.

In the building phase, we chop the target image into tiles of some multiple of it's aspect ratio. Then each of those patches are scanned against the source images in the data blob to determine the best matching images. This matching process is done in the L*A*B* colorspace since it appears to give better color matches.

## Caveats

* Only images with matching (or rotated) aspect ratio can be used as tiles. That is if the target image is 4:3, source images of aspect ratio 4:3 or 3:4 will be considered. 3:4 images when used as tiles in this example will be rotated 90 degrees.
* The tool *always* picks an image for *every* tile. This means that if there are no good fits in the source image collection, you may wind up with a sub-optimal matching tile.
* Source images with arbitrarily large aspect ratio multiple (multiple 4 and 3 for 4:3 to get 12), are rejected as source images. This is to keep the size of the index reasonable.
