# Notification Examples

This document provides complete examples for creating notifications through different channels.

## Email Notification

Send an email notification with a templated HTML message.

```json
{
  "title": "Welcome to our platform",
  "content": "Hi! Welcome to our amazing platform. We're glad to have you here.",
  "channel_name": "email",
  "meta": {
    "to": "user@example.com",
    "subject": "Welcome!",
    "template": "titled"
  }
}
```

**Required meta fields:**
- `to`: Valid email address
- `subject`: Email subject (optional)
- `template`: Template name - "titled" or "plain" (optional, defaults to "titled")

---

## SMS Notification

Send an SMS notification (max 160 characters).

```json
{
  "title": "Account Verification",
  "content": "Your verification code is: 123456. Valid for 10 minutes.",
  "channel_name": "sms",
  "meta": {
    "phone": "+1234567890",
    "send_date": "2024-10-21"
  }
}
```

**Required meta fields:**
- `phone`: Valid phone number in E.164 format (e.g., +1234567890)
- `send_date`: Date in YYYY-MM-DD format

---

## Push Notification

Send a push notification to a mobile device.

```json
{
  "title": "New Message",
  "content": "You have a new message from John",
  "channel_name": "push",
  "meta": {
    "token": "device_token_xyz123",
    "platform": "android",
    "data": "{\"message_id\":\"123\",\"chat_id\":\"456\"}"
  }
}
```

**Required meta fields:**
- `token`: Device token for push notification
- `platform`: Platform type - "android" or "ios"
- `data`: Additional data as JSON string (optional)

---

## Testing Flow

1. **Login** to get a JWT token:
   ```bash
   curl -X POST http://localhost:8080/login \
     -H "Content-Type: application/json" \
     -d '{"email":"test@test.com","password":"password123"}'
   ```

2. **Create Email Notification**:
   ```bash
   curl -X POST http://localhost:8080/notifications \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer YOUR_TOKEN_HERE" \
     -d '{
       "title": "Welcome",
       "content": "Welcome message",
       "channel_name": "email",
       "meta": {
         "to": "user@example.com",
         "subject": "Welcome!"
       }
     }'
   ```

3. **Create SMS Notification**:
   ```bash
   curl -X POST http://localhost:8080/notifications \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer YOUR_TOKEN_HERE" \
     -d '{
       "title": "Verification",
       "content": "Your code: 123456",
       "channel_name": "sms",
       "meta": {
         "phone": "+1234567890",
         "send_date": "2024-10-21"
       }
     }'
   ```

4. **Create Push Notification**:
   ```bash
   curl -X POST http://localhost:8080/notifications \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer YOUR_TOKEN_HERE" \
     -d '{
       "title": "New Message",
       "content": "You have a message",
       "channel_name": "push",
       "meta": {
         "token": "device_token",
         "platform": "android"
       }
     }'
   ```

