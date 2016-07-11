Go PHoto PRocessor
==================

**Dependencies**

*For make*

 * sudo apt-get install libgif-dev
 * sudo apt-get install libmagickwand-dev

*For use*

 * sudo apt-get install libmagickcore5
 * sudo apt-get install libgif4

**Input parameters:**

 * size (NxM) — for resize
 * bestfit (0|1) — for best fit cropping
 * watermark (0|1) — add watermark

**Configuration**

[http]

 * addr = :8187 # where to bind server
 * keep_alive = 300 # keep alive timeout
 * access_log = false # disable access log

[proxy]

 * url = http://strg.kolesa.kz/ # where to find images
 * timeout = 1000 # how long to wait for upstream response

[image]

 * background = 255,255,255 # background color for non-bestfit cropping

[watermark]

 * color_threshold = 127,127,127 # threshold for choosing black/white watermark image
 * file_black_big = watermark-b-b.png # filename for black big watermark
 * file_black_small = watermark-b-s.png # filename for black small watermark
 * file_white_big = watermark-w-b.png # filename for white big watermark
 * file_white_small = watermark-w-s.png # filename for white small watermark
 * margin = 20 # margin for watermark in pixels
 * path = /path/to/directory/ # path ro directory containing watermark files
 * size_threshold = 280x210 # threshold for choosing big/small watermark image
