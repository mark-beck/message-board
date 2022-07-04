use anyhow::bail;
use anyhow::Result;
use serde::Deserialize;
use serde::Serialize;
use tracing::warn;

use crate::config::ImageServiceConfig;

#[derive(Debug, Clone)]
pub struct ImageService {
    url: String,
}

impl ImageService {
    pub fn new(config: ImageServiceConfig) -> Self {
        Self {
            url: config.url
        }
    }

    #[tracing::instrument(name = "upload_image")]
    pub async fn upload_image(&self, image_data: String) -> Result<String> {
        #[derive(Debug, Serialize, Deserialize)]
        struct UploadRequest {
            pub data: String,
        }

        #[derive(Debug, Serialize, Deserialize)]
        struct UploadResponse {
            pub id: String,
        }

        let client = awc::Client::default();
        let mut response = match client
            .post(&format!("http://{}/upload", self.url))
            .send_json(&UploadRequest { data: image_data })
            .await
        {
            Ok(r) => r,
            Err(err) => {
                warn!("{:?}", err);
                bail!("Error sending request to image_service")
            }
        };

        let id = match response.json::<UploadResponse>().await {
            Ok(r) => r,
            Err(err) => {
                warn!("{:?}", err);
                bail!("Error decoding response image_service")
            }
        }.id;

        Ok(id)
    }
}
