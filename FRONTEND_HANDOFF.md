# Frontend Handoff Guide

Welcome to the frontend integration phase of the Diagnostyka-App! This document provides essential information to help you connect your frontend application to the backend API seamlessly.

## 1. CORS & Authentication

The backend uses **HTTP-only cookies** for authentication. This is highly secure but requires specific configurations on the frontend:

- **Credentials**: Every request (including login, profile updates, etc.) MUST be sent with credentials enabled.
  - If using `fetch`: Add `credentials: 'include'` to your fetch options.
  - If using `axios`: Add `withCredentials: true` to your axios instance.
- **Tokens**: You do not need to manually read or attach access tokens to an `Authorization` header. The browser will handle the `access_token` and `refresh_token` cookies automatically.

## 2. API Documentation

Comprehensive details for all endpoints can be found in the `API_DOCUMENTATION.md` file. Key highlights include:
- All profile updates (Patient and Doctor) use the `PATCH` method, allowing you to send only the fields that have changed.
- The Patient's PESEL is returned masked (e.g., `XXXXXXX1234`). If you send a profile update containing the masked PESEL, the backend safely ignores it.
- To view the full PESEL, you must use the specific `/api/patient/profile/pesel` endpoint and prompt the user for their password.

## 3. Environment Setup

For local development, ensure the backend is running. By default, it runs on `http://localhost:8080`.

1. Copy `.env.example` to `.env` (or coordinate with the backend dev for the required keys).
2. Run the server using `go run server.go`.

> **Note on CORS:** Currently, the backend CORS is configured to allow requests from `http://localhost:4200`. If your frontend runs on a different port, please notify the backend developer to add your origin to the `AllowOrigins` list in `server.go`.

## 4. Error Handling

- **400 Bad Request**: Invalid JSON body or validation errors (e.g., invalid email format, weak password, invalid PESEL).
- **401 Unauthorized**: Missing or expired session cookies.
- **403 Forbidden**: Access denied. Most commonly, this returns an `{"error": "profile_incomplete"}` message indicating the user needs to complete their profile setup before accessing certain endpoints.

## 5. Dates and Timestamps

- In the active appointments list, the `date` key represents the booking creation time.
- In the appointment history list, the `date` key represents the time the doctor completed the appointment and submitted the results.
