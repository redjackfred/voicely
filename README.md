# Voicely (替聲) - MVP

> Voicely is a zero-UI AI voice agent that calls traditional, non-digital stores for you. Built with Go & Next.js, it bridges the gap for users with phone anxiety by handling real-time negotiations, placing orders, and dynamically crowdsourcing store knowledge.

## 💡 Core Concept (Zero-UI)

Many local businesses (bentos, traditional clinics, breakfast shops) still rely exclusively on phone calls. Voicely acts as a proxy: users trigger an order via a shortcut or voice command, and our AI agent handles the actual phone call, negotiates alternatives if items are sold out, and returns the final status.

## 🏗 Architecture & Data Flow

Voicely uses an event-driven architecture to handle real-time SIP calls and low-latency API responses.
