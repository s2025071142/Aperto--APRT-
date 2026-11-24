# Aperto (APTO)

Bitcoin’s 21M supply cap establishes it as "digital gold," but its real-world transactional utility remains throttled by layer-1 throughput constraints, rigid smart contract limitations, and long-term state bloat. For long-term viability, it requires structural protocol upgrades—such as tail emissions for security subsidies, advanced Turing-complete programmability, and lightweight node synchronization.

---

## Technical Features Implemented In This Code

- **RandomX PoW Engine (`pow.go`):** Implements Monero's native mining algorithm to ensure strict CPU decentralization and block ASIC centralization.
- **Continuous Tail Emission (`mint.go`):** Features a smooth, continuous fractional decay per block that locks into a perpetual floor of **0.005 coins** forever.
- **Dynamic Block Sizing (`abci.go`):** Automatically resizes block execution limits based on transaction traffic trends to maintain ultra-low fees under a penny.
- **EVM Execution Layer (`go.mod`):** Native compatibility for Solidity smart contracts, decentralized applications (dApps), and standard Web3 wallets like MetaMask.
- **Dual-Mode Lightweight Node Storage (`state.go`):** Supports interchangeable **UTXO Commitment Mode** (default on mobile/light devices) and **Pruned Node Mode** (default on desktop/PC) to prevent hardware centralization while maintaining full validation.

---

## Node Storage & Sync Modes

Node operators can freely switch between UTXO commitments and historical block pruning regardless of their device:

| Node Mode | Default Platform | Storage Footprint | How It Works |
| :--- | :--- | :--- | :--- |
| **UTXO Commitment Mode (`utxo`)** | **Mobile / Light Devices** *(Default)* | **~Kilobytes (KB)** | Verifies active balances instantly using cryptographic UTXO accumulators without downloading historical blocks or storing the full state set. |
| **Pruned Node Mode (`pruned`)** | **Desktop / PC** *(Default)* | **~10–30 GB** | Downloads and validates new blocks in real time, then automatically deletes old spent transaction history while keeping the active state set locally. |
| **Archival Mode (`archival`)** | **Servers / Explorers** *(Optional)* | **Full History** | Retains every block and historical state transition from Genesis. |

---

## Build and Launch Guide

Ensure you have Go 1.21+ installed on your operating system, then execute these commands in your terminal:

```bash
# 1. Download project dependencies and compile the blockchain binary
go mod tidy
go build -o mycustoml1d main.go

# 2. Initialize genesis records and operational configuration states
./mycustoml1d init local-node-test --chain-id l1-custom-network

# 3. Fire up your mining validator node (auto-selects 'utxo' on mobile or 'pruned' on PC)
./mycustoml1d start

# Optional: Explicitly override the default node storage mode on any device
./mycustoml1d start --storage-mode=utxo
./mycustoml1d start --storage-mode=pruned --prune-keep-recent=1000
```

## License
Distributed freely under the open-source MIT License terms.
