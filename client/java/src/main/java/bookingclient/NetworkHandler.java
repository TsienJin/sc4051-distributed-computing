package bookingclient;

import java.io.IOException;
import java.net.*;
import java.nio.ByteBuffer;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Comparator;
import java.util.List;
import java.util.*; // Import Map and HashMap

/**
 * Handles network communication (UDP sending and receiving) for the booking client.
 * Implements logic for reliable request-response handling with ACKs, retries,
 * timeout management, and filtering of responses based on message IDs.
 */
public class NetworkHandler {
    // Network Configuration
    private static final String HOST = "100.105.193.66"; // Replace with actual host if needed
    static final int UDP_PORT = 8765;
    private static final int TIMEOUT_MS = 1000; // Timeout for each receive attempt
    private static final int MAX_PACKET_SIZE = 1024; // Define max expected packet size

    // Retry Configuration
    private int maxRetries = 3; // Default max overall retries (timeouts) for a request
    private int maxRetriesAfterError = 10; // Default max retries (timeouts) *after* receiving an error response for the current request

    // Network Components
    private DatagramSocket socket;
    private static InetAddress address;
    private static final PacketUnmarshaller unmarshaller = new PacketUnmarshaller();
    // Note: PacketMarshaller is used statically within methods

    // State Variables for the current sendPacketWithAckAndResend operation
    private int overallRetries;
    private int retriesSinceLastError;
    private boolean errorReceived; // Has an error response *for the current request* been received?
    private Map<Integer, Packet> receivedSuccessParts; // Success parts *for the current request*
    private List<Packet> errorResponsePackets; // Error response *for the current request*
    private boolean isRequestAcknowledged; // Has the server ACKed our *current request* packet?
    private boolean receivedAnySuccessPart; // Have we received any success part *for the current request*?
    private int expectedSuccessTotalPackets; // Total success packets expected *for the current request*
    private UUID currentRequestMessageId; // The Message ID of the request currently being processed

    // --- Configuration Setters ---

    /** Sets the DatagramSocket to be used. Useful for testing or custom socket configurations. */
    public void setSocket(DatagramSocket socket) {
        this.socket = socket;
    }

    /** Sets the maximum number of overall timeouts before giving up on a request. */
    public void setMaxRetries(int maxRetries) {
        this.maxRetries = maxRetries;
    }

    /** Sets the maximum number of timeouts to wait *after* receiving an error response, allowing time for a potential overriding success response. */
    public void setMaxRetriesAfterError(int maxRetriesAfterError) {
        this.maxRetriesAfterError = maxRetriesAfterError;
    }

    // --- Initialization ---

    /**
     * Initializes the UDP socket and resolves the server address.
     * Must be called before sending packets unless auto-initialization occurs.
     * @throws RuntimeException if initialization fails.
     */
    public void networkClient(){
        try {
            // OS chooses an available local port. Specify a port if needed: new DatagramSocket(LOCAL_PORT);
            socket = new DatagramSocket();
            address = InetAddress.getByName(HOST);
            Debugger.log("Network client initialized. Socket bound to local port: " + socket.getLocalPort() + ", Server: " + address.getHostAddress());
        } catch (IOException e) {
            // Log detailed error
            Debugger.log("FATAL: Failed to initialize network client. Error: " + e.getMessage());
            e.printStackTrace(); // Print stack trace for debugging
            // Wrap in RuntimeException to avoid forcing checked exceptions up the call stack
            // where immediate handling might not be feasible. Consider a custom exception.
            throw new RuntimeException("Failed to initialize network client", e);
        }
    }

    /** Ensures the network socket and address are initialized. Called internally. */
    private void initializeNetworkIfNeeded() {
        if (socket == null || address == null) {
            Debugger.log("Network components not initialized. Attempting auto-initialization...");
            networkClient(); // Try to initialize
            if (socket == null) { // Check again
                throw new IllegalStateException("NetworkHandler not initialized and auto-initialization failed. Cannot proceed.");
            }
        }
    }

    /** Resets the state variables used for a single sendPacketWithAckAndResend operation. */
    private void initializeSendState() {
        overallRetries = 0;
        retriesSinceLastError = 0;
        errorReceived = false;
        receivedSuccessParts = new HashMap<>();
        errorResponsePackets = new ArrayList<>(); // Will hold max 1 error packet for the current request
        isRequestAcknowledged = false;
        receivedAnySuccessPart = false;
        expectedSuccessTotalPackets = -1;
        currentRequestMessageId = null; // Reset message ID for the new operation
        Debugger.log("Send state initialized for new request.");
    }

    // --- Core Send/Receive Logic ---

    /**
     * Sends a request packet and reliably handles the response sequence:
     * - Extracts the request's Message ID from its header.
     * - Sends the request.
     * - Waits for responses (ACK, RESPONSE, REQUEST_RESEND).
     * - Extracts original request ID from ACK/RESPONSE payloads for relevance check.
     * - ACKs all received RESPONSE packets (to stop server retries).
     * - Filters packets based on whether they correspond to the `currentRequestMessageId`.
     * - Handles success (potentially fragmented) and error responses for the current request.
     * - Implements timeouts, retries (with backoff), and error tolerance windows.
     *
     * @param requestPacketBytes The byte array representing the marshalled request packet to send.
     * @return A List of Packet objects representing the final response for the request:
     *         - Complete Success: Contains all sorted parts of the success response.
     *         - Confirmed Error: Contains the single error packet received (if no success override occurred).
     *         - Partial Success (Timeout): Contains the incomplete list of success parts received before timeout.
     *         - Timeout (No Response): Contains an empty list.
     * @throws IOException If a network error occurs (sending/receiving), initialization fails,
     *                     or the outgoing packet's Message ID cannot be determined.
     */
    public List<Packet> sendPacketWithAckAndResend(byte[] requestPacketBytes) throws IOException {
        initializeNetworkIfNeeded();
        initializeSendState(); // Reset state for this specific request

        // --- Crucial: Get Message ID from the outgoing request packet header ---
        try {
            // Use the unmarshaller to parse the header of the packet *we just built*.
            Packet requestPacketHeaderView = unmarshaller.unmarshalResponse(requestPacketBytes); // Adjust method if needed
            this.currentRequestMessageId = requestPacketHeaderView.messageId();

            if (this.currentRequestMessageId == null) {
                throw new IOException("Failed to extract valid Message ID (UUID was null) from outgoing request packet header.");
            }
            Debugger.log("Current Request Message ID (from outgoing packet header): " + this.currentRequestMessageId);
        } catch (Exception e) {
            Debugger.log("FATAL: Could not determine Message ID of outgoing packet. Cannot proceed reliably. Error: " + e.getMessage());
            throw new IOException("Failed to extract Message ID from outgoing request packet: " + e.getMessage(), e);
        }
        // --- End Message ID retrieval ---

        // Prepare the DatagramPacket for sending
        DatagramPacket outgoingDatagram = new DatagramPacket(requestPacketBytes, requestPacketBytes.length, address, UDP_PORT);

        Debugger.log("Sending packet (request)... Length: " + requestPacketBytes.length + ", Hex: " + PacketMarshaller.bytesToHex(requestPacketBytes));
        socket.send(outgoingDatagram);
        Debugger.log("Request packet sent to " + address.getHostAddress() + ":" + UDP_PORT + " [MsgID: " + this.currentRequestMessageId + "]");
        socket.setSoTimeout(TIMEOUT_MS);

        // Main receive loop
        while (true) {
            try {
                byte[] packetBuffer = new byte[MAX_PACKET_SIZE];
                DatagramPacket incomingDatagram = new DatagramPacket(packetBuffer, packetBuffer.length);
                socket.receive(incomingDatagram); // Wait for a packet

                byte[] receivedData = Arrays.copyOf(incomingDatagram.getData(), incomingDatagram.getLength());
                Debugger.log("Received raw packet: Length=" + receivedData.length + ", Hex=" + PacketMarshaller.bytesToHex(receivedData));

                Packet receivedPacket;
                try {
                    // Unmarshal the received data
                    receivedPacket = unmarshaller.unmarshalResponse(receivedData);
                } catch (Exception e) {
                    Debugger.log("Error unmarshalling received packet: " + e.getMessage() + ". Ignoring packet.");
                    socket.setSoTimeout(TIMEOUT_MS); // Reset timeout before continuing
                    continue; // Skip malformed packet
                }

                // Process the valid received packet. Handles filtering, ACKing, state updates.
                List<Packet> result = processReceivedPacket(receivedPacket, outgoingDatagram);
                if (result != null) {
                    return result;
                }

            } catch (SocketTimeoutException e) {
                List<Packet> result = handleTimeout(outgoingDatagram);
                if (result != null) {
                    return result;
                }
                // Otherwise, null means retry logic was executed, continue the loop

            } catch (IOException ioEx) {
                Debugger.log("IOException during receive/process for [MsgID: " + this.currentRequestMessageId + "]: " + ioEx.getMessage());
                throw ioEx; // Rethrow IOExceptions
            } catch (Exception genEx) {
                Debugger.log("Unexpected error during receive/process loop for [MsgID: " + this.currentRequestMessageId + "]: " + genEx.getMessage());
                throw new RuntimeException("Unexpected error in receive loop: " + genEx.getMessage(), genEx);
            }
        } // End of while(true) loop
    }


    /**
     * Processes a successfully received and unmarshalled packet.
     * - Determines packet type.
     * - Extracts relevant Message IDs (response header ID, original request ID from payload).
     * - Sends ACKs for *all* valid RESPONSE packets immediately.
     * - Filters processing based on whether the packet relates to the `currentRequestMessageId`.
     * - Updates internal state (acknowledgement, received parts, errors).
     *
     * @param receivedPacket The unmarshalled packet from the server.
     * @param originalOutgoingDatagram The original request datagram (needed for resends).
     * @return List of packets if the operation is now complete (e.g., full success received), otherwise null.
     * @throws IOException If sending an ACK or resending the request fails.
     */
    private List<Packet> processReceivedPacket(Packet receivedPacket, DatagramPacket originalOutgoingDatagram) throws IOException {
        PacketType receivedType = PacketType.fromCode(receivedPacket.messageType());
        UUID responseHeaderMessageId = receivedPacket.messageId(); // ID of the packet received (can be null if unmarshalling failed partially)
        UUID originalRequestIdFromPayload = null; // Populated for relevant types
        boolean isRelevantToCurrentRequest = false;

        if (receivedType == null) {
            Debugger.log("Received packet with unknown type code: " + receivedPacket.messageType() + " [RespMsgID: " + responseHeaderMessageId + "]. Ignoring.");
            socket.setSoTimeout(TIMEOUT_MS);
            return null;
        }

        // Determine relevance based on packet type and appropriate ID comparison
        switch (receivedType) {
            case ACK:
                // For ACKs, relevance is determined by the *original request ID* embedded in its payload.
                originalRequestIdFromPayload = extractOriginalRequestIdFromPayload(receivedPacket.payload());
                if (originalRequestIdFromPayload == null) {
                    Debugger.log("WARNING: Could not extract original request ID from ACK payload [RespMsgID: " + responseHeaderMessageId + ", Pkt#: " + receivedPacket.packetNumber() + "]. Cannot determine relevance.");
                    isRelevantToCurrentRequest = false;
                } else {
                    isRelevantToCurrentRequest = this.currentRequestMessageId != null && this.currentRequestMessageId.equals(originalRequestIdFromPayload);
                }
                break; // Proceed to relevance check below

            case REQUEST_RESEND:
                // For REQUEST_RESEND, the relevant ID is the one in *its* header,
                isRelevantToCurrentRequest = this.currentRequestMessageId != null && this.currentRequestMessageId.equals(responseHeaderMessageId);
                break; // Proceed to relevance check below

            case RESPONSE:
                // For RESPONSE, relevance is determined by the *original request ID* embedded in its payload.
                originalRequestIdFromPayload = extractOriginalRequestIdFromPayload(receivedPacket.payload());
                if (originalRequestIdFromPayload == null) {
                    Debugger.log("WARNING: Could not extract original request ID from RESPONSE payload [RespMsgID: " + responseHeaderMessageId + ", Pkt#: " + receivedPacket.packetNumber() + "]. Cannot determine relevance.");
                    isRelevantToCurrentRequest = false;
                } else {
                    isRelevantToCurrentRequest = this.currentRequestMessageId != null && this.currentRequestMessageId.equals(originalRequestIdFromPayload);
                }

                // **Crucial: ACK every valid RESPONSE immediately, regardless of relevance**
                try {
                    sendAckForResponse(receivedPacket); // Uses receivedPacket's header info
                } catch (IOException ackEx) {
                    Debugger.log("WARNING: Failed to send ACK for RESPONSE packet #" + receivedPacket.packetNumber()
                            + " [RespMsgID: " + responseHeaderMessageId + ", OrigReqID in Payload: " + originalRequestIdFromPayload + "]. Error: " + ackEx.getMessage());
                } catch (Exception e) {
                    Debugger.log("WARNING: Failed to construct/marshal ACK for RESPONSE packet #" + receivedPacket.packetNumber()
                            + " [RespMsgID: " + responseHeaderMessageId + ", OrigReqID in Payload: " + originalRequestIdFromPayload + "]. Error: " + e.getMessage());
                }
                break; // Proceed to relevance check below

            default:
                Debugger.log("Received packet of unexpected type: " + receivedType + " [RespMsgID: " + responseHeaderMessageId + "]. Ignoring.");
                socket.setSoTimeout(TIMEOUT_MS);
                return null;
        }

        // --- Process based on relevance ---

        if (!isRelevantToCurrentRequest) {
            String logOrigId = (originalRequestIdFromPayload != null) ? originalRequestIdFromPayload.toString() : "N/A or Unextracted";
            Debugger.log("Received stale " + receivedType + " packet [RespMsgID: " + responseHeaderMessageId
                    + ", Pkt#: " + receivedPacket.packetNumber()
                    + ", OrigReqID in Payload: " + logOrigId // Log ID from payload if available
                    + "]. Does not match current request [Expected OrigReqID: " + this.currentRequestMessageId + "]. Ignoring payload/action.");
            socket.setSoTimeout(TIMEOUT_MS); // Reset timeout even if ignoring
            return null; // Ignore this packet's content/action for the current operation
        }

        // --- Packet IS relevant to the current request ---
        String logOrigIdRelevant = (originalRequestIdFromPayload != null) ? originalRequestIdFromPayload.toString() : this.currentRequestMessageId.toString(); // Use extracted or expected ID
        Debugger.log("Processing relevant " + receivedType + " packet for current request [Expected/Matched OrigReqID: " + logOrigIdRelevant
                + ", RespMsgID: " + responseHeaderMessageId // Header ID of the received packet
                + ", Pkt#: " + receivedPacket.packetNumber() + "]");

        switch (receivedType) {
            case ACK:
                // Relevance already confirmed using payload ID
                handleAckPacket(); // Handles ACK for the original request we sent
                socket.setSoTimeout(TIMEOUT_MS); // Reset timeout
                break;

            case RESPONSE:
                // Relevance already confirmed using payload ID, ACK already sent
                if (receivedPacket.getStatus() == 200) {
                    boolean complete = handleSuccessResponsePacket(receivedPacket);
                    if (complete) {
                        return getSortedSuccessPackets(); // COMPLETE SUCCESS for this request
                    }
                } else {
                    handleErrorResponsePacket(receivedPacket);
                }
                socket.setSoTimeout(TIMEOUT_MS); // Reset timeout after processing relevant response
                break;

            case REQUEST_RESEND:
                // Relevance already confirmed using header ID
                handleRequestResendPacket(originalOutgoingDatagram);
                // handleRequestResendPacket resets its own timeout
                break;

            // Default case handled before relevance check
        }

        return null; // Indicate processing occurred, but operation is not yet complete
    }

    // --- Helper Methods (mostly unchanged, operate on current request's state) ---

    /**
     * Extracts the original request's UUID from the payload of a response OR ACK packet.
     * Assumes the UUID is the first 16 bytes of the payload.
     *
     * @param payload The payload byte array from the received packet.
     * @return The extracted UUID, or null if the payload is too short or null or if extraction fails.
     */
    private UUID extractOriginalRequestIdFromPayload(byte[] payload) {
        if (payload == null || payload.length < 16) {
            // Debugger.log("Payload too short (" + (payload == null ? "null" : payload.length) + " bytes) to extract original request ID.");
            return null;
        }
        try {
            ByteBuffer pb = ByteBuffer.wrap(payload);
            byte[] originalIdBytes = new byte[16];
            pb.get(originalIdBytes); // Read the first 16 bytes
            // Assuming PacketUnmarshaller.fromByteArray exists and is accessible
            UUID extractedId = PacketUnmarshaller.fromByteArray(originalIdBytes);
            return extractedId;
        } catch (Exception e) {
            // Log specific error during extraction
            Debugger.log("Error extracting original request ID from payload bytes " + PacketMarshaller.bytesToHex(Arrays.copyOf(payload, 16)) + ": " + e.getMessage());
            return null;
        }
    }

    /** Handles receiving an ACK *for the current request* (relevance checked via payload ID). Sets the acknowledged flag. */
    private void handleAckPacket() {
        if (!isRequestAcknowledged) {
            Debugger.log("Received relevant ACK (payload matches current request) [OrigReqID: " + this.currentRequestMessageId + "]. Marking request as acknowledged.");
            isRequestAcknowledged = true;
        } else {
            Debugger.log("Received duplicate relevant ACK for [OrigReqID: " + this.currentRequestMessageId + "]. Ignoring.");
        }
    }

    /**
     * Handles a success (status 200) RESPONSE packet *for the current request*.
     * (Relevance checked via payload ID). Stores packet, updates state, checks completion.
     */
    private boolean handleSuccessResponsePacket(Packet receivedPacket) {
        int packetNum = receivedPacket.packetNumber();

        if (!receivedSuccessParts.containsKey(packetNum)) {
            receivedSuccessParts.put(packetNum, receivedPacket);
            receivedAnySuccessPart = true;
            Debugger.log("Stored relevant success RESPONSE part #" + packetNum + " for [OrigReqID: " + this.currentRequestMessageId + "]");

            // Update expected total packet count
            if (expectedSuccessTotalPackets == -1 || expectedSuccessTotalPackets != receivedPacket.totalPackets()) {
                if (expectedSuccessTotalPackets != -1) {
                    Debugger.log("Warning: Total packet count changed from " + expectedSuccessTotalPackets + " to " + receivedPacket.totalPackets() + " for [OrigReqID: " + this.currentRequestMessageId + "]");
                }
                expectedSuccessTotalPackets = receivedPacket.totalPackets();
                Debugger.log("Expecting " + expectedSuccessTotalPackets + " success packets total for [OrigReqID: " + this.currentRequestMessageId + "]");
            }

            // Check if all parts received
            if (isSuccessComplete()) {
                Debugger.log("All " + expectedSuccessTotalPackets + " expected successful response packets received for [OrigReqID: " + this.currentRequestMessageId + "].");
                return true; // Complete!
            }
        } else {
            Debugger.log("Received duplicate relevant success RESPONSE packet #" + packetNum + " for [OrigReqID: " + this.currentRequestMessageId + "]. Ignoring content, ACK already sent.");
        }

        // Reset error state if success arrives after error
        if (errorReceived) {
            Debugger.log("Received a relevant success part after an error part for [OrigReqID: " + this.currentRequestMessageId + "]. Prioritizing success. Resetting error state.");
            errorReceived = false;
            errorResponsePackets.clear();
            retriesSinceLastError = 0;
        }
        return false; // Not complete yet
    }

    /**
     * Handles an error (status != 200) RESPONSE packet *for the current request*.
     * (Relevance checked via payload ID). Stores the first error unless success parts seen.
     */
    private void handleErrorResponsePacket(Packet receivedPacket) {
        Debugger.log("Handling relevant ERROR RESPONSE packet [OrigReqID: " + this.currentRequestMessageId
                + ", RespMsgID: " + receivedPacket.messageId()
                + ", Pkt#: " + receivedPacket.packetNumber() + "/" + receivedPacket.totalPackets()
                + ", Status: " + receivedPacket.getStatus() + "]");

        if (receivedPacket.totalPackets() != 1) {
            Debugger.log("Warning: Received multi-packet error response (total: "
                    + receivedPacket.totalPackets() + ") for [OrigReqID: " + this.currentRequestMessageId + "]. Treating as single error event.");
        }

        // Store error only if no success parts seen yet AND no previous error stored
        if (!receivedAnySuccessPart && errorResponsePackets.isEmpty()) {
            Debugger.log("Storing first encountered relevant error response (Status: " + receivedPacket.getStatus() + ") for [OrigReqID: " + this.currentRequestMessageId + "]. Starting limited wait.");
            errorResponsePackets.add(receivedPacket);
            errorReceived = true;
            retriesSinceLastError = 0;
        } else if (!errorResponsePackets.isEmpty()) {
            Debugger.log("Received another relevant error response (Status: " + receivedPacket.getStatus() + ") for [OrigReqID: " + this.currentRequestMessageId + "]. Keeping first error stored.");
        } else {
            Debugger.log("Received relevant error response (Status: " + receivedPacket.getStatus() + ") for [OrigReqID: " + this.currentRequestMessageId + "], but success parts already received. Ignoring this error.");
        }
    }

    /** Handles a REQUEST_RESEND packet *for the current request* (relevance checked via header ID). */
    private void handleRequestResendPacket(DatagramPacket originalOutgoingDatagram) throws IOException {
        // The header ID of REQUEST_RESEND matches our original request ID
        Debugger.log("Received relevant REQUEST_RESEND (header matches current request) for [OrigReqID: " + this.currentRequestMessageId + "]. Resending original request packet.");
        socket.send(originalOutgoingDatagram);
        socket.setSoTimeout(TIMEOUT_MS); // Reset timeout after resend
    }

    /** Handles SocketTimeoutException, checks limits, determines final response or performs retry. */
    private List<Packet> handleTimeout(DatagramPacket originalOutgoingDatagram) throws IOException {
        overallRetries++;
        Debugger.log("Timeout waiting for response for [OrigReqID: " + this.currentRequestMessageId + "], overall retry attempt " + overallRetries);

        // Check Post-Error Timeout Limit
        if (errorReceived) {
            retriesSinceLastError++;
            Debugger.log("Timeout occurred after receiving error for [OrigReqID: " + this.currentRequestMessageId + "]. Post-error retry " + retriesSinceLastError + "/" + this.maxRetriesAfterError);
            if (this.maxRetriesAfterError >= 0 && retriesSinceLastError > this.maxRetriesAfterError) {
                Debugger.log("Max retries after error exceeded for [OrigReqID: " + this.currentRequestMessageId + "]. Finalizing response.");
                if (isSuccessComplete()) {
                    Debugger.log("Complete success response received before post-error timeout limit for [OrigReqID: " + this.currentRequestMessageId + "]. Returning success.");
                    return getSortedSuccessPackets();
                }
                Debugger.log("Returning stored error response for [OrigReqID: " + this.currentRequestMessageId + "] due to post-error timeout.");
                return errorResponsePackets;
            }
        }

        // Check Overall Timeout Limit
        if (this.maxRetries >= 0 && overallRetries > this.maxRetries) {
            Debugger.log("Overall max retries (" + this.maxRetries + ") exceeded for [OrigReqID: " + this.currentRequestMessageId + "]. Determining final response...");
            return determineFinalResponse();
        }

        // Perform Retry Action
        performRetryAction(originalOutgoingDatagram); // Resend request or log wait
        backoffDelay(overallRetries);
        socket.setSoTimeout(TIMEOUT_MS);

        return null; // Indicate retry performed, loop should continue
    }

    /** Determines the final response list when a timeout limit is hit. */
    private List<Packet> determineFinalResponse() {
        if (isSuccessComplete()) {
            Debugger.log("Finalizing: Returning complete success response for [OrigReqID: " + this.currentRequestMessageId + "].");
            return getSortedSuccessPackets();
        } else if (!errorResponsePackets.isEmpty()) {
            Debugger.log("Finalizing: Returning stored error response for [OrigReqID: " + this.currentRequestMessageId + "].");
            return errorResponsePackets;
        } else if (receivedAnySuccessPart) {
            Debugger.log("Finalizing: Returning potentially incomplete success response for [OrigReqID: " + this.currentRequestMessageId + "].");
            return getSortedSuccessPackets();
        } else {
            Debugger.log("Finalizing: Returning empty list (timeout, no relevant response received) for [OrigReqID: " + this.currentRequestMessageId + "].");
            return new ArrayList<>();
        }
    }

    /** Checks if a complete success response *for the current request* has been received. */
    private boolean isSuccessComplete() {
        return receivedAnySuccessPart && expectedSuccessTotalPackets > 0 && receivedSuccessParts.size() == expectedSuccessTotalPackets;
    }

    /** Returns the collected success packets *for the current request*, sorted. */
    private List<Packet> getSortedSuccessPackets() {
        List<Packet> successPackets = new ArrayList<>(receivedSuccessParts.values());
        successPackets.sort(Comparator.comparingInt(Packet::packetNumber));
        return successPackets;
    }

    /** Performs the retry action: resend original request if not ACKed. */
    private void performRetryAction(DatagramPacket originalOutgoingDatagram) throws IOException {
        if (!isRequestAcknowledged) {
            socket.send(originalOutgoingDatagram);
            Debugger.log("Resending original request packet for [OrigReqID: " + this.currentRequestMessageId + "] (overall attempt " + overallRetries + ")");
        } else {
            Debugger.log("Original request ACKed for [OrigReqID: " + this.currentRequestMessageId + "]. Continuing wait for full response (overall attempt " + overallRetries + ")");
        }
    }

    // --- Low-Level Helpers ---

    /** Constructs and sends an ACK packet for a received RESPONSE packet. */
    private void sendAckForResponse(Packet responsePacket) throws IOException {
        UUID responseMsgId = responsePacket.messageId();
        byte responsePktNum = responsePacket.packetNumber();

        // ACK Payload: ID of the packet being ACKed (16 bytes) + Packet number of packet being ACKed (1 byte)
        ByteBuffer payload = ByteBuffer.allocate(16 + 1);
        try {
            if (responseMsgId == null) {
                throw new IllegalArgumentException("Cannot create ACK payload with null response message ID.");
            }
            // Use the ID and packet number from the *response* header
            payload.put(PacketMarshaller.UUIDtoByteArray(responseMsgId));
            payload.put(responsePktNum);
        } catch (Exception e) {
            String errorMsg = "Error constructing ACK payload for [RespMsgID: " + responseMsgId + ", Pkt#: " + responsePktNum + "]: " + e.getMessage();
            Debugger.log(errorMsg);
            throw new IOException(errorMsg, e);
        }

        byte[] ackPayloadBytes = payload.array();

        // Marshal the ACK packet
        byte[] ackPacketData;
        try {
            ackPacketData = PacketMarshaller.marshalPacket(
                    PacketType.ACK.getCode(),
                    (byte) 0x01,  // Packet Number for ACK itself
                    (byte) 0x01,  // Total Packets for ACK itself
                    false, false, // Flags (AckReq=false, Fragment=false)
                    ackPayloadBytes
            );
        } catch (Exception e) {
            String errorMsg = "Error marshalling ACK packet for [RespMsgID: " + responseMsgId + ", Pkt#: " + responsePktNum + "]: " + e.getMessage();
            Debugger.log(errorMsg);
            throw new IOException(errorMsg, e);
        }

        // Send ACK
        DatagramPacket ackPacket = new DatagramPacket(ackPacketData, ackPacketData.length, address, UDP_PORT);
        socket.send(ackPacket);
        Debugger.log("Sent ACK for response packet [RespMsgID: " + responseMsgId + ", Pkt#: " + responsePktNum + "]");
    }

    /** Exponential backoff delay with jitter. */
    private void backoffDelay(int retryCount) {
        try {
            int magnitude = Math.min(retryCount, 5);
            long baseDelay = 20 * (long) Math.pow(2, magnitude);
            long jitter = (long) (baseDelay * (Math.random() * 0.4 - 0.2));
            long delay = Math.max(50, baseDelay + jitter);
            Debugger.log("Delaying for " + delay + "ms before retry " + (retryCount + 1) + "...");
            Thread.sleep(delay);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            Debugger.log("Backoff delay interrupted.");
        }
    }

    /** Closes the UDP socket. */
    public void closeSocket() {
        if (socket != null && !socket.isClosed()) {
            socket.close();
            Debugger.log("Socket closed.");
        }
    }

    // --- Monitoring Method (Updated for Payload ID Check & Enhanced Step 2 Logging) ---
    public List<Packet> sendMonitorFacilityPacket(byte[] packet, int ttl) throws IOException {
        initializeNetworkIfNeeded(); // Ensure socket is ready

        // --- Get Monitor Request ID from its header ---
        UUID monitorRequestId;
        try {
            // Assuming unmarshaller can parse the header part of the request packet
            Packet requestPacketHeaderView = unmarshaller.unmarshalResponse(packet); // Adjust if needed
            monitorRequestId = requestPacketHeaderView.messageId();
            if (monitorRequestId == null) {
                throw new IOException("Failed to extract valid Message ID (UUID was null) from outgoing monitor request packet header.");
            }
            Debugger.log("Monitor Request Message ID: " + monitorRequestId);
        } catch (Exception e) {
            Debugger.log("FATAL: Could not determine message ID of outgoing monitor packet. Error: " + e.getMessage());
            throw new IOException("Failed to extract Message ID from outgoing monitor request packet: " + e.getMessage(), e);
        }
        // ---

        int connectionRetries = 0;
        long startTime = System.currentTimeMillis();
        long ttlEndTime = startTime + ttl * 1000;
        List<Packet> responsePackets = new ArrayList<>(); // Store monitored updates
        Set<String> receivedUpdateKeys = new HashSet<>(); // Track received "RespHeaderID:PacketNum" for monitor updates

        Debugger.log("Sending monitoring request [MsgID: " + monitorRequestId + "]... TTL: " + ttl + "s.");
        DatagramPacket datagramPacket = new DatagramPacket(packet, packet.length, address, UDP_PORT);
        try {
            socket.send(datagramPacket);
        } catch (IOException sendEx) {
            Debugger.log("ERROR: Failed to send initial monitor request: " + sendEx.getMessage());
            throw sendEx; // Rethrow if initial send fails
        }
        Debugger.log("Monitoring request sent. Waiting for initial ACK or Response...");

        boolean connectionLive = false;

        // Step 1: Initial check for liveness (ACK or first RESPONSE related to our monitor request)
        // Use a temporary higher retry count for this crucial phase if needed, or ensure default is high enough
        int initialMaxRetries = this.maxRetries; // Use the configured maxRetries
        // Note: Consider setting a dedicated higher retry limit specifically for monitor connection
        // if this.maxRetries is often low for other operations. For now, uses the global setting.

        while (!connectionLive && (connectionRetries <= initialMaxRetries || initialMaxRetries < 0) ) { // Allow negative maxRetries for infinite
            // If maxRetries is 0, loop runs once. If negative, runs indefinitely until connectionLive or error.
            if (initialMaxRetries >= 0 && connectionRetries > initialMaxRetries) {
                Debugger.log("Max retries (" + initialMaxRetries + ") exceeded for initial monitor check.");
                break; // Exit loop, connectionLive is still false
            }

            socket.setSoTimeout(TIMEOUT_MS); // Use standard timeout for initial check attempts
            try {
                byte[] packetBuffer = new byte[MAX_PACKET_SIZE];
                DatagramPacket recvPacket = new DatagramPacket(packetBuffer, packetBuffer.length);

                socket.receive(recvPacket); // Wait for ACK or Response

                byte[] receivedData = Arrays.copyOf(recvPacket.getData(), recvPacket.getLength());

                Packet receivedPacket;
                try {
                    receivedPacket = unmarshaller.unmarshalResponse(receivedData);
                } catch (Exception unmarshalEx) {
                    Debugger.log("Initial Check: Error unmarshalling packet: " + unmarshalEx.getMessage() + ". Raw Data: " + PacketMarshaller.bytesToHex(receivedData) +". Ignoring.");
                    continue; // Ignore malformed packet
                }

                PacketType receivedType = PacketType.fromCode(receivedPacket.messageType());
                UUID respHeaderId = receivedPacket.messageId();
                UUID origReqIdFromPayload = null;

                if (receivedType == PacketType.ACK || receivedType == PacketType.RESPONSE) {
                    origReqIdFromPayload = extractOriginalRequestIdFromPayload(receivedPacket.payload());
                }

                Debugger.log("Initial Check Received: Type=" + receivedType + ", RespMsgID=" + respHeaderId + ", OrigReqIDPayload=" + (origReqIdFromPayload != null ? origReqIdFromPayload : "null/NA"));

                boolean relevant = monitorRequestId.equals(origReqIdFromPayload);

                if (receivedType == PacketType.ACK && relevant) {
                    Debugger.log("Relevant ACK received for monitor request [MsgID: " + monitorRequestId + "]. Connection confirmed live.");
                    connectionLive = true; // SUCCESS!
                } else if (receivedType == PacketType.RESPONSE) {
                    // ACK *every* valid RESPONSE immediately
                    try { sendAckForResponse(receivedPacket); } catch (Exception e) { Debugger.log("WARN: Failed ACK during initial monitor check: " + e.getMessage()); }

                    if (relevant) {
                        Debugger.log("Initial relevant RESPONSE received for monitor request [MsgID: " + monitorRequestId + "]. Connection confirmed live.");
                        connectionLive = true; // SUCCESS!
                        // Store this first update, track its key
                        String packetKey = respHeaderId + ":" + receivedPacket.packetNumber();
                        if(receivedUpdateKeys.add(packetKey)) {
                            responsePackets.add(receivedPacket);
                            Debugger.log("Stored initial monitored update packet #" + receivedPacket.packetNumber());
                        } else {
                            Debugger.log("WARN: Initial relevant response was a duplicate? Key: " + packetKey);
                        }
                    } else {
                        Debugger.log("Received RESPONSE for different request ("+origReqIdFromPayload+") during initial check. Ignoring payload, ACK sent.");
                    }
                } else if (!relevant && receivedType == PacketType.ACK){
                    Debugger.log("Received stale/irrelevant ACK [OrigReqID in Payload: "+origReqIdFromPayload+"]. Ignoring for liveness check.");
                }
                else {
                    Debugger.log("Ignoring unexpected/irrelevant packet type " + receivedType + " during initial check.");
                }
            } catch (SocketTimeoutException e) {
                connectionRetries++;
                Debugger.log("Timeout waiting for initial monitor ACK/Response (retry " + connectionRetries + "/" + (initialMaxRetries < 0 ? "inf" : initialMaxRetries) + ") for [MsgID: " + monitorRequestId +"]");
                // Check if retries are exhausted *before* attempting resend
                if (initialMaxRetries >= 0 && connectionRetries > initialMaxRetries) {
                    Debugger.log("Max retries reached for initial monitor check after timeout.");
                    break; // Exit loop
                }
                // Retry only if limit not hit
                backoffDelay(connectionRetries);
                try {
                    socket.send(datagramPacket); // Resend monitor request
                    Debugger.log("Resent monitoring request [MsgID: " + monitorRequestId + "]");
                } catch (IOException sendEx) {
                    Debugger.log("ERROR: Failed to resend monitor request on retry " + connectionRetries + ": " + sendEx.getMessage());
                    // If resend fails, likely network is down, probably should give up.
                    throw new IOException("Failed to resend monitor request after timeout.", sendEx);
                }
            } catch (IOException ioEx) {
                Debugger.log("IOException during initial monitor check receive: " + ioEx.getMessage());
                // Depending on error, maybe wait and retry or rethrow
                connectionRetries++; // Count as a retry attempt maybe?
                if (initialMaxRetries >= 0 && connectionRetries > initialMaxRetries) break; // Prevent infinite loop on persistent IO error
                backoffDelay(connectionRetries); // Wait before potentially trying again
            }
            // Catching generic Exception for unmarshalling is now handled inside the main try

        } // End Step 1 while loop

        // Check connectionLive and throw exception *after* the loop finishes
        if (!connectionLive) {
            throw new IOException("Failed to establish live connection for monitoring [MsgID: " + monitorRequestId + "] after " + connectionRetries + " attempts (max_retries=" + initialMaxRetries + "). Check network connectivity and server status/drop rate.");
        }


        // Step 2: Monitor for updates within TTL
        Debugger.log("Connection confirmed. Monitoring updates for " + ttl + " seconds... [OrigReqID: " + monitorRequestId + "]");

        while (System.currentTimeMillis() < ttlEndTime) {
            long remainingTime = ttlEndTime - System.currentTimeMillis();
            if (remainingTime <= 0) {
                Debugger.log("Monitoring TTL loop terminating: Remaining time <= 0.");
                break;
            }

            // Use min 100ms timeout to avoid busy-waiting
            int currentTimeout = (int) Math.max(100, Math.min(remainingTime, TIMEOUT_MS));
            socket.setSoTimeout(currentTimeout);
            // Debugger.log("Monitor Step 2: Set SO_TIMEOUT to " + currentTimeout + "ms");

            try {
                byte[] packetBuffer = new byte[MAX_PACKET_SIZE];
                DatagramPacket recvPacket = new DatagramPacket(packetBuffer, packetBuffer.length);

                // -----> Try to receive <-----
                socket.receive(recvPacket);
                // -----> If receive() returns without exception, a packet WAS received <-----

                int receivedLength = recvPacket.getLength();
                String senderAddr = recvPacket.getAddress().getHostAddress() + ":" + recvPacket.getPort();
                Debugger.log("Monitor Step 2: Successfully received datagram. Length=" + receivedLength + ", Sender=" + senderAddr); // Log receipt confirmation

                byte[] receivedData = Arrays.copyOf(recvPacket.getData(), receivedLength);

                Packet receivedPacket;
                try {
                    receivedPacket = unmarshaller.unmarshalResponse(receivedData);
                    Debugger.log("Monitor Step 2: Unmarshalled packet: Type=" + PacketType.fromCode(receivedPacket.messageType())
                            + ", RespMsgID=" + receivedPacket.messageId()
                            + ", Pkt#=" + receivedPacket.packetNumber() + "/" + receivedPacket.totalPackets());
                } catch (Exception e) {
                    Debugger.log("Monitor Step 2: Error unmarshalling received packet (Length=" + receivedLength + ", Sender=" + senderAddr + "): " + e.getMessage()
                            + ". Raw Data: " + PacketMarshaller.bytesToHex(receivedData) + ". Ignoring.");
                    continue; // Skip malformed packet
                }

                PacketType receivedType = PacketType.fromCode(receivedPacket.messageType());
                UUID respHeaderId = receivedPacket.messageId();

                if (receivedType == PacketType.RESPONSE) {
                    // Always ACK received RESPONSE packets
                    try { sendAckForResponse(receivedPacket); }
                    catch (IOException ackIoEx) { Debugger.log("WARN: IOException sending ACK during monitoring for [RespMsgID: " + respHeaderId + ", Pkt#: " + receivedPacket.packetNumber() + "]: " + ackIoEx.getMessage()); }
                    catch (Exception ackEx) { Debugger.log("WARN: Error preparing/sending ACK during monitoring for [RespMsgID: " + respHeaderId + ", Pkt#: " + receivedPacket.packetNumber() + "]: " + ackEx.getMessage()); }

                    // Log monitored payload content
                    try {
                        String monitoredContent = PacketUnmarshaller.monitoredPayloadBytesToString(receivedPacket.payload());
                        if (monitoredContent != null) {
                            Debugger.log("  Monitored Payload Content: " + monitoredContent);
                        } else {
                            // It's okay if payload isn't string decodable, might be binary update
                            Debugger.log("  Monitored update payload is not simple text or format unrecognized.");
                        }
                    } catch (Exception payloadEx) {
                        Debugger.log("  Error reading monitored payload content: " + payloadEx.getMessage());
                    }

                    // Store if unique (based on *response* header ID and packet number)
                    String packetKey = respHeaderId + ":" + receivedPacket.packetNumber();
                    if (receivedUpdateKeys.add(packetKey)) {
                        Debugger.log("Stored unique monitored update: [RespMsgID: " + respHeaderId + ", Pkt#: " + receivedPacket.packetNumber() + "/" + receivedPacket.totalPackets() + "]");
                        responsePackets.add(receivedPacket);
                    } else {
                        Debugger.log("Received duplicate monitored packet [RespMsgID: " + respHeaderId + ", Pkt#: " + receivedPacket.packetNumber() + "]. Ignoring content, ACK sent.");
                    }

                } else {
                    // Ignore other packet types during active monitoring
                    Debugger.log("Ignoring packet type " + receivedType + " during monitoring phase [RespMsgID: " + respHeaderId + "].");
                }
            } catch (SocketTimeoutException e) {
                // This is NORMAL if no updates arrive or packets are lost
                Debugger.log("Timeout while monitoring (waiting for updates)... Remaining time: " + remainingTime + "ms");
            } catch (IOException ioEx) {
                Debugger.log("IOException during monitoring loop receive: " + ioEx.getMessage());
                // Consider if this indicates a fatal problem for monitoring, maybe break?
            } catch (Exception genEx) {
                Debugger.log("Unexpected error during monitoring loop processing: " + genEx.getMessage());
                // Log and continue monitoring
            }
        } // End Step 2 while loop

        // Step 3: TTL expired
        Debugger.log("Monitoring TTL expired for [OrigReqID: " + monitorRequestId + "]. Received " + responsePackets.size() + " unique response packets.");
        responsePackets.sort(Comparator.comparing(Packet::messageId).thenComparingInt(Packet::packetNumber));
        return responsePackets;
    }
}