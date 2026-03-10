# Diagnostyka-App
**System backendowy wspierający proces diagnostyki medycznej, zarządzanie pakietami badań oraz automatyzację przydziału pracy lekarzom.**
## Architektura Technologiczna
- Język: Golang
- Framework Web: Gin 
- ORM: GORM
- Baza danych: PostgreSQL

## Główne Funkcjonalności
### Moduł Pacjenta

- Zarządzanie kontem: Rejestracja oraz bezpieczne logowanie do systemu.
- Rezerwacje: Możliwość zapisu na konkretne pakiety badań diagnostycznych.
- Dostarczanie wyników: Automatyczna wysyłka wiadomości e-mail z kodem QR umożliwiającym dostęp do wyników badań.

### Moduł Lekarza

- Inteligentny Kolejkowanie: Algorytm przydzielający badania do lekarzy na podstawie ich aktualnego obciążenia pracą.
- Zarządzanie badaniami: Akceptacja i przetwarzanie badań przypisanych bezpośrednio do konkretnego specjalisty.
- Panel Analityczny (Dashboard): Pełny wgląd w status wszystkich badań w systemie wraz z informacją o osobach odpowiedzialnych za ich realizację.