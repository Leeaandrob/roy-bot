# roy-bot

Bot to control and help in automated tasks

## Commands

### To transcoder a video

$ /transcoder_job <tenant_schema> <video_id> <video_name> <video_height> <enviroment>

#### Example:

$ /transcoder_job ry649df1c6c1 5ccc3eed5af8c65be9e19f10 video3 1080p release


### To verify the status of the job transcoder

$ /transcoder_status <tenant_schema> <video_id> <video_name> <video_height>

#### Example:

$ /transcoder_job_status ry649df1c6c1 5ccc39f25af8c65be9e19f0e 360p


### To get the actual status of transcoder

$ /transcoder_status
