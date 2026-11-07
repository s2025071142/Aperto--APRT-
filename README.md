# Aperto(Aperto)

Bitcoin’s 21M supply cap establishes it as "digital gold," but its real-world transactional utility remains throttled by layer-1 throughput constraints and rigid smart contract limitations. For long-term viability, it requires structural protocol upgrades—such as tail emissions for security subsidies and advanced Turing-complete programmability.

---

## Technical Features Implemented In This Code

- **RandomX PoW Engine (`pow.go`):** Implements Monero's native mining algorithm to ensure strict CPU decentralization and block ASIC centralization.
- **Continuous Tail Emission (`mint.go`):** Features a smooth, continuous fractional decay per block that locks into a perpetual floor of **0.005 coins** forever.
- **Dynamic Block Sizing (`abci.go`):** Automatically resizes block execution limits based on transaction traffic trends to maintain ultra-low fees under a penny.
- **EVM Execution Layer (`go.mod`):** Native compatibility for Solidity smart contracts, decentralized applications (dApps), and standard Web3 wallets like MetaMask.

## Build and Launch Guide

Ensure you have Go 1.21+ installed on your operating system, then execute these commands in your terminal:

```bash
# 1. Download project dependencies and compile the blockchain binary
go mod tidy
go build -o mycustoml1d main.go

# 2. Initialize genesis records and operational configuration states
./mycustoml1d init local-node-test --chain-id l1-custom-network

# 3. Fire up your mining validator node and open local EVM RPC interfaces
./mycustoml1d start
```

## License
Distributed freely under the open-source MIT License terms.
