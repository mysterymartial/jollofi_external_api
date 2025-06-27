Jollfi Gaming API
The Jollfi Gaming API provides endpoints for staking and managing games on the Sui blockchain. It supports optional API key authentication and is designed for public access.
Base Information

Base URL: https://api.jollfi.com (or http://localhost:8080 for local development)
Authentication: Optional API key via X-API-Key header or api_key query parameter (public-jollfi-api-key-2025)
Rate Limit: 100 requests per minute per IP
Content-Type: application/json for POST requests

Endpoints
GET /health
Checks the API's health status.

Request:curl https://api.jollfi.com/health


Response:{
  "success": true,
  "status": "healthy",
  "service": "jollfi-gaming-api",
  "timestamp": 1622134567,
  "version": "1.0.0"
}



GET /api/v1/info
Provides API information and available endpoints.

Request:curl https://api.jollfi.com/api/v1/info


Response:{
  "success": true,
  "info": {
    "service": "Jollfi Gaming API",
    "version": "1.0.0",
    "environment": "development",
    "network": "https://fullnode.testnet.sui.io:443",
    "module": "jollfi_wallet",
    "endpoints": {
      "stake": "POST /api/v1/games/stake",
      "pay_winner": "POST /api/v1/games/pay_winner",
      "stake_history": "GET /api/v1/games/stakes/:address",
      "game_history": "GET /api/v1/games/history/:address",
      "stats": "GET /api/v1/games/stats",
      "health": "GET /health"
    }
  }
}



GET /api/v1/status
Returns API status and uptime.

Request:curl https://api.jollfi.com/api/v1/status


Response:{
  "success": true,
  "service": "jollfi-gaming-api",
  "version": "1.0.0",
  "environment": "development",
  "timestamp": 1622134567,
  "uptime": "1h2m3s"
}



POST /api/v1/games/stake
Creates a stake on the Sui blockchain.

Request:curl -X POST -H "Content-Type: application/json" \
     -H "X-API-Key: public-jollfi-api-key-2025" \
     -d '{
          "requester_coin_id": "0x123",
          "accepter_coin_id": "0x456",
          "requester_address": "0x1234567890abcdef1234567890abcdef12345678",
          "accepter_address": "0xabcdef1234567890abcdef1234567890abcdef12",
          "stake_amount": 100
        }' https://api.jollfi.com/api/v1/games/stake


Without API Key:curl -X POST -H "Content-Type: application/json" \
     -d '{
          "requester_coin_id": "0x123",
          "accepter_coin_id": "0x456",
          "requester_address": "0x1234567890abcdef1234567890abcdef12345678",
          "accepter_address": "0xabcdef1234567890abcdef1234567890abcdef12",
          "stake_amount": 100
        }' https://api.jollfi.com/api/v1/games/stake




Response:{
  "success": true,
  "transaction_digest": "tx_1234567890",
  "message": "Stake successful. 10% fee deducted from each player by blockchain."
}


Errors:
400: Invalid request format, missing fields, or negative stake amount.
500: Blockchain or database error.



POST /api/v1/games/pay_winner
Processes payment to the winner on the Sui blockchain.

Request:curl -X POST -H "Content-Type: application/json" \
     -H "X-API-Key: public-jollfi-api-key-2025" \
     -d '{
          "requester_address": "0x1234567890abcdef1234567890abcdef12345678",
          "accepter_address": "0xabcdef1234567890abcdef1234567890abcdef12",
          "stake_amount": 100,
          "requester_score": 10,
          "accepter_score": 5
        }' https://api.jollfi.com/api/v1/games/pay_winner


Response:{
  "success": true,
  "transaction_digest": "tx_9876543210",
  "message": "Winner payment processed by blockchain with all fees calculated automatically."
}


Errors:
400: Invalid request format or missing fields.
500: Blockchain or database error.



GET /api/v1/games/stakes/:address
Retrieves stake history for a Sui address (up to 50 records).

Request:curl https://api.jollfi.com/api/v1/games/stakes/0x1234567890abcdef1234567890abcdef12345678


Response:{
  "success": true,
  "stakes": [
    {
      "requester_coin_id": "0x123",
      "accepter_coin_id": "0x456",
      "requester_address": "0x1234567890abcdef1234567890abcdef12345678",
      "accepter_address": "0xabcdef1234567890abcdef1234567890abcdef12",
      "stake_amount": 100,
      "status": "completed",
      "timestamp": 1622134567,
      "transaction_hash": "tx_1234567890"
    }
  ],
  "count": 1
}


Errors:
400: Invalid address format.
500: Database error.



GET /api/v1/games/history/:address
Retrieves game history for a Sui address (up to 50 records).

Request:curl https://api.jollfi.com/api/v1/games/history/0x1234567890abcdef1234567890abcdef12345678


Response:{
  "success": true,
  "games": [
    {
      "requester_address": "0x1234567890abcdef1234567890abcdef12345678",
      "accepter_address": "0xabcdef1234567890abcdef1234567890abcdef12",
      "requester_score": 10,
      "accepter_score": 5,
      "stake_amount": 100,
      "timestamp": 1622134567,
      "transaction_hash": "tx_9876543210"
    }
  ],
  "count": 1
}


Errors:
400: Invalid address format.
500: Database error.



GET /api/v1/games/stats
Placeholder for game statistics.

Request:curl https://api.jollfi.com/api/v1/games/stats


Response:{
  "success": true,
  "message": "Game stats endpoint - to be implemented",
  "data": {
    "total_games": 0,
    "total_stakes": 0,
    "active_players": 0,
    "last_updated": 1622134567
  }
}
