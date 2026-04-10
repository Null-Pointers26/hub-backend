# Hub Backend

Jednoduchý backend pro Hub frontend.

Program funguje jako reverzní proxy:
- Vrací seznam her pro frontend přes API.
- Přeposílá požadavky na konkrétní hru podle URL cesty.
- V produkci umí HTTPS s automatickým certifikátem (Let's Encrypt přes autocert).

## Co backend dělá

- GET /api/games vrací seznam dostupných her.
- /games/{id}/... přeposílá provoz na Target dané hry.
- Ostatní cesty (např. /) přeposílá na frontend službu.

Data her jsou aktuálně uložena jen v paměti procesu.

## API z pohledu Hub frontendu

### 1) Načtení seznamu her

Endpoint:
- GET /api/games
- Plné URL: https://DOMAIN/api/games (kde DOMAIN je konfigurována v ENV)

Příklad z frontendu:

```js
const res = await fetch('/api/games');
if (!res.ok) throw new Error('Nepodařilo se načíst hry');
const games = await res.json();
```

Očekávaný tvar každé položky:

```json
{
  "id": "chess",
  "name": "Chess",
  "target": "http://game-chess:3000",
  "icon": "♟",
  "author": "Team",
  "image": "https://..."
}
```

### 2) Otevření konkrétní hry

Frontend může odkazovat na cestu:
- https://DOMAIN/games/{id}
- https://DOMAIN/games/{id}/nějaká/další/cesta

Backend automaticky:
- najde hru podle id,
- přepošle požadavek na target dané hry,
- zachová zbytek cesty za id.

Příklad:
- https://DOMAIN/games/chess/match/123
- přepošle se na target hry chess jako /match/123

### Poznámka o relativních cestách

V frontendu je možné používat relativní cesty (začínají /) bez znalosti konkrétní domény:

```js
const res = await fetch('/api/games');
```

Prohlížeč automaticky interpretuje `/api/games` vůči aktuální doméně. Příklad:
- Návštěvník přišel na https://mujhub.cz
- fetch('/api/games') se přeloží na https://mujhub.cz/api/games
- Na localhostu: http://localhost:8080/api/games

Takže frontend nemusíte znát konkrétní doménu - relativní cesty se vždy řeší vůči doméně, na které je aplikace dostupná.

## Konfigurace přes ENV

Pro nasazení je potřeba vytvořit .env soubor s potřebnými proměnnými.

Nejdůležitější proměnné:
- APP_ENV: development nebo production
- DOMAIN: doména pro certifikát v produkci
- CERT_EMAIL: kontakt pro Let's Encrypt
- FRONTEND_TARGET: URL frontend služby
- DEV_LISTEN_ADDR: adresa pro lokální HTTP (default :8080)
- HTTP_ADDR: produkční HTTP (default :80)
- HTTPS_ADDR: produkční HTTPS (default :443)
- CERT_CACHE_DIR: složka pro cache certifikátů

## TODO

- udělat persistenci dat
- správu her
