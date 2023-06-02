import os
import sys
import time
from pathlib import Path
# dapr
# ffmpeg
import ffmpeg

video_dir = Path.cwd() / 'video'
tmp_dir = Path.cwd() / 'tmp'
os.makedirs(str(tmp_dir), exist_ok=True)

for f in os.listdir(str(video_dir)):
    p = video_dir / f
    if 'empty' in f:
        continue
    print(f)
    t0 = time.time()
    try:
        # probe = ffmpeg.probe('video/earth_120.avi')
        probe = ffmpeg.probe(str(p))
    except ffmpeg.Error as e:
        print('stdout: %s, stderr: %s' %(
            e.stdout, e.stderr,
        ))
        sys.exit(1)
    t1 = time.time()
    print('ffprobe time = %.1f ms' %(1000*(t1-t0)))
    # print(probe)
    print(probe['format']['format_name'])
    duration = float(probe['format']['duration'])
    video_stream = next((stream for stream in probe['streams'] if stream['codec_type'] == 'video'), None)
    # if video_stream is None:
    #     print('No video spotted')
    # print(video_stream)
    if video_stream != None:
        width = video_stream['width']
    else:
        width = 0
    targ_width = 0
    if width == 1920:
        targ_width = 1280
    elif width == 1280:
        targ_width = 640
    # scale
    if targ_width != 0:
        targ_f = '%d-%s' %(targ_width, f)
        # test scale
        t0 = time.time()  
        try: 
            (
                ffmpeg
                .input(str(p))
                .filter('scale', width, -1)
                .output(str(tmp_dir / targ_f), preset='slow', crf=18)
                .overwrite_output()
                # .run_async(pipe_stdout=True, pipe_stderr=True)
                .run(capture_stdout=True, capture_stderr=True)
            ) 
        except ffmpeg.Error as e:
            out = e.stdout.decode()
            err = e.stderr.decode()
            print('FFmpeg scale err: %s, out: %s' %(err, out))

        t1 = time.time()
        print('ffmpeg scale time = %.1f ms' %(1000*(t1-t0)))

        # test thumbnail
        t0 = time.time()
        ss = min(0.1, duration/10)
        targ_f = 'tn-%s.jpeg' %f
        try:
            (
                ffmpeg
                .input(str(p), ss=ss)
                .output(str(tmp_dir / targ_f), vframes=1, format='image2', vcodec='mjpeg')
                .overwrite_output()
                # .run_async(pipe_stdout=True, pipe_stderr=True)
                .run(capture_stdout=True, capture_stderr=True)
            )
        except ffmpeg.Error as e:
            out = e.stdout.decode()
            err = e.stderr.decode()
            print('FFmpeg thumnbail err: %s, out: %s' %( err, out))
        t1 = time.time()
        print('ffmpeg thumbnail time = %.1f ms' %(1000*(t1-t0)))
    print('---------------------')