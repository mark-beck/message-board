use crate::config::Config;
use lettre::transport::smtp::response::Response;
use lettre::{Message, SmtpTransport, Transport};
use tracing::{trace, warn};

#[allow(dead_code)]
pub struct Mailer {
    config: Config,
}

impl Mailer {
    pub const fn with_config(config: Config) -> Self {
        Self {
            config,
        }
    }

    pub async fn send_mail(
        &self,
        address: &str,
        subject: &str,
        body: &str,
    ) -> Result<Response, anyhow::Error> {
        trace!("send_mail");
        let email = Message::builder()
            .from(self.config.mail.sender.parse()?)
            .to(address.parse()?)
            .subject(subject)
            .body(String::from(body))?;

        // Create TLS transport on port 465

        let sender = SmtpTransport::builder_dangerous(&self.config.mail.server).build();
        // let sender = SmtpTransport::builder_dangerous("localhost").build();
        // Send the email via remote relay
        let result = sender.send(&email);
        warn!("{:?}", result);
        if let Err(ref e) = result {
            warn!("{}", e);
        }
        result.map_err(Into::into)
    }
}
