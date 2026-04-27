# Diagnostyka-App API Documentation

This document outlines the available REST API endpoints for the Diagnostyka-App application. 

## Base URL
All API requests should be made relative to the base URL, which defaults to `http://localhost:8080` in the local development environment.

## Authentication Overview
The API uses cookie-based authentication. 
- Upon successful login, the server sets `access_token` and `refresh_token` HTTP-only cookies.
- These cookies are automatically attached by the browser on subsequent requests.
- Ensure the frontend application is configured to include credentials (e.g., `credentials: 'include'` in fetch, or `withCredentials: true` in Axios) for CORS requests.

---

## 1. Authentication Endpoints

### 1.1. Register User
- **URL**: `/auth/register`
- **Method**: `POST`
- **Access**: Public
- **Description**: Registers a new user account.
- **Request Body**:
  ```json
  {
    "email": "user@example.com",
    "password": "securepassword123",
    "role": "patient", // "patient" or "doctor"
    "specialization": "Cardiology" // Optional, required only if role is "doctor"
  }
  ```
- **Responses**:
  - `200 OK`: Account created successfully.
  - `400 Bad Request`: Invalid input or email already exists.

### 1.2. Login
- **URL**: `/auth/login`
- **Method**: `POST`
- **Access**: Public
- **Description**: Authenticates a user and sets HTTP-only session cookies.
- **Request Body**:
  ```json
  {
    "email": "user@example.com",
    "password": "securepassword123",
    "role": "patient"
  }
  ```
- **Responses**:
  - `200 OK`: Authenticated successfully.
  - `401 Unauthorized`: Invalid credentials.

### 1.3. Refresh Token
- **URL**: `/auth/refresh`
- **Method**: `POST`
- **Access**: Public (requires valid `refresh_token` cookie)
- **Description**: Issues a new `access_token` if the current one has expired.
- **Responses**:
  - `200 OK`: Token refreshed successfully.
  - `401 Unauthorized`: Invalid or expired refresh token.

### 1.4. Logout
- **URL**: `/auth/logout`
- **Method**: `POST`
- **Access**: Public
- **Description**: Invalidates the current session and clears authentication cookies.
- **Responses**:
  - `200 OK`: Logged out successfully.

---

## 2. General Endpoints

### 2.1. Get Current User Status
- **URL**: `/api/me`
- **Method**: `GET`
- **Access**: Authenticated (Any Role)
- **Description**: Retrieves the current session status, role, and whether the profile setup is complete.
- **Responses**:
  - `200 OK`: Returns session data.
    ```json
    {
      "authenticated": true,
      "message": "You are currently logged in",
      "profile_complete": true,
      "role": "patient",
      "user_id": 1
    }
    ```

### 2.2. List Available Examinations
- **URL**: `/api/examinations`
- **Method**: `GET`
- **Access**: Authenticated (Any Role)
- **Description**: Retrieves a list of all available medical examinations in the system.
- **Responses**:
  - `200 OK`: Returns an array of examination objects.

---

## 3. Patient Endpoints

### 3.1. Get Patient Profile
- **URL**: `/api/patient/profile`
- **Method**: `GET`
- **Access**: Patient
- **Description**: Retrieves the patient's profile details. The PESEL number is returned in a masked format for privacy.
- **Responses**:
  - `200 OK`: Returns profile data.
    ```json
    {
      "email": "user@example.com",
      "first_name": "John",
      "id": 1,
      "last_name": "Doe",
      "pesel": "XXXXXXX1234"
    }
    ```

### 3.2. Update Patient Profile
- **URL**: `/api/patient/profile`
- **Method**: `PATCH`
- **Access**: Patient
- **Description**: Updates the patient's profile. Supports partial updates. If a masked PESEL is submitted, it is ignored to prevent data corruption.
- **Request Body** (All fields optional):
  ```json
  {
    "first_name": "John",
    "last_name": "Doe",
    "pesel": "12345678901"
  }
  ```
- **Responses**:
  - `200 OK`: Profile updated successfully.
  - `400 Bad Request`: Invalid input data or invalid PESEL checksum.

### 3.3. Reveal Full PESEL
- **URL**: `/api/patient/profile/pesel`
- **Method**: `POST`
- **Access**: Patient
- **Description**: Retrieves the patient's unmasked PESEL number. Requires password confirmation.
- **Request Body**:
  ```json
  {
    "password": "current_account_password"
  }
  ```
- **Responses**:
  - `200 OK`: Returns decrypted PESEL.
    ```json
    {
      "pesel": "12345678901"
    }
    ```
  - `401 Unauthorized`: Incorrect password.

### 3.4. Get Active Appointments
- **URL**: `/api/patient/appointments/active`
- **Method**: `GET`
- **Access**: Patient (Requires Profile Complete)
- **Description**: Retrieves pending, incomplete appointments for the patient.
- **Responses**:
  - `200 OK`: Returns an array of active appointments.

### 3.5. Get Appointment History
- **URL**: `/api/patient/appointments/history`
- **Method**: `GET`
- **Access**: Patient (Requires Profile Complete)
- **Description**: Retrieves completed appointments and their corresponding results.
- **Responses**:
  - `200 OK`: Returns an array of completed appointments.

### 3.6. Book Appointment
- **URL**: `/api/patient/book`
- **Method**: `POST`
- **Access**: Patient (Requires Profile Complete)
- **Description**: Schedules a new appointment for a specific examination.
- **Request Body**:
  ```json
  {
    "examination_id": 1
  }
  ```
- **Responses**:
  - `201 Created`: Appointment booked successfully. Returns appointment details.
  - `400 Bad Request`: Patient already has an active appointment for this examination.
  - `503 Service Unavailable`: No doctors are available for the selected examination.

### 3.7. Cancel Appointment
- **URL**: `/api/patient/appointments/:id`
- **Method**: `DELETE`
- **Access**: Patient (Requires Profile Complete)
- **Description**: Cancels a pending appointment.
- **Responses**:
  - `200 OK`: Appointment cancelled successfully.
  - `400 Bad Request`: Cannot cancel a completed appointment or invalid ID format.
  - `403 Forbidden`: Attempting to cancel another user's appointment.

---

## 4. Doctor Endpoints

### 4.1. Get Doctor Profile
- **URL**: `/api/doctor/profile`
- **Method**: `GET`
- **Access**: Doctor
- **Description**: Retrieves the doctor's profile details and managed examinations.
- **Responses**:
  - `200 OK`: Returns profile data.

### 4.2. Update Doctor Profile
- **URL**: `/api/doctor/profile`
- **Method**: `PATCH`
- **Access**: Doctor
- **Description**: Updates the doctor's profile. Supports partial updates. Returns the updated profile object.
- **Request Body** (All fields optional):
  ```json
  {
    "first_name": "Gregory",
    "last_name": "House",
    "specialization": "Diagnostician",
    "examination_ids": [1, 2, 3]
  }
  ```
- **Responses**:
  - `200 OK`: Profile updated successfully.

### 4.3. Get Active Schedule
- **URL**: `/api/doctor/schedule/active`
- **Method**: `GET`
- **Access**: Doctor (Requires Profile Complete)
- **Description**: Retrieves pending, incomplete appointments assigned to the doctor.
- **Responses**:
  - `200 OK`: Returns an array of active appointments.

### 4.4. Get Schedule History
- **URL**: `/api/doctor/schedule/history`
- **Method**: `GET`
- **Access**: Doctor (Requires Profile Complete)
- **Description**: Retrieves completed appointments performed by the doctor.
- **Responses**:
  - `200 OK`: Returns an array of completed appointments.

### 4.5. Complete Appointment
- **URL**: `/api/doctor/appointments/:id/complete`
- **Method**: `POST`
- **Access**: Doctor (Requires Profile Complete)
- **Description**: Finalizes an appointment by submitting diagnostic results. Triggers an email notification to the patient.
- **Request Body**:
  ```json
  {
    "result": "Patient is healthy.",
    "notes": "Follow up in 6 months." // Optional
  }
  ```
- **Responses**:
  - `200 OK`: Appointment completed successfully.
  - `400 Bad Request`: Appointment already completed or invalid ID format.
  - `403 Forbidden`: Doctor is not assigned to this appointment.

---

## 5. Public Results Endpoints

### 5.1. View Public Results
- **URL**: `/api/results/:hash`
- **Method**: `GET`
- **Access**: Public
- **Description**: Retrieves the details and results of a completed appointment using its unique hash.
- **Responses**:
  - `200 OK`: Returns result data.
  - `404 Not Found`: Invalid hash or appointment not yet completed.

### 5.2. View QR Code
- **URL**: `/api/results/:hash/qr`
- **Method**: `GET`
- **Access**: Public
- **Description**: Returns a PNG image of the QR code linking to the results page.
- **Responses**:
  - `200 OK`: Returns image/png data.
  - `404 Not Found`: Invalid hash.
