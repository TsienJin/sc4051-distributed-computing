package bookingclient;
import java.io.IOException;
import java.net.*;
import java.nio.ByteBuffer;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Comparator;
import java.util.List;

public class NetworkHandler {
    private static final String HOST = "100.105.193.66";
    private static final int UDP_PORT = 8765;
    private static final int TIMEOUT_MS = 2000;
    private static final int MAX_RETRIES = 1;
    private static DatagramSocket socket;
    private static InetAddress address;
    private static final PacketUnmarshaller unmarshaller = new PacketUnmarshaller();

    public void networkClient(){
        try {
            socket = new DatagramSocket();
            address = InetAddress.getByName("100.105.193.66");
        } catch (IOException e) {
            e.printStackTrace();
        }
    }

    public byte[] sendPacketWithAck(byte[] packet) throws IOException {
        //This function sends a packet and waits for an acknowledgement and response.
        DatagramPacket DatagramPacket = new DatagramPacket(packet, packet.length, address, UDP_PORT);
        byte[] ackBuffer = new byte[1024];
        DatagramPacket ackPacket = new DatagramPacket(ackBuffer, ackBuffer.length);
        int retries = 0;
        while (retries < MAX_RETRIES) {
            try {
                System.out.println(bytesToHex(DatagramPacket.getData()));
                socket.send(DatagramPacket);
                socket.receive(ackPacket);
                System.out.println("ack time" + bytesToHex(ackPacket.getData()));
                socket.receive(ackPacket);
                System.out.println("second time" + bytesToHex(ackPacket.getData()));
                return ackPacket.getData();
            } catch (IOException e) {
                retries++;
                System.out.println("Failed to send packet, retrying...");
            }
        }

        throw new IOException("Failed to send packet after " + MAX_RETRIES + " retry attempts.");

    }
    public List<Packet> sendMonitorFacilityPacket(byte[] packet, int ttl) throws IOException {
        int retries = 0;
        long startTime = System.currentTimeMillis();
        long ttlEndTime = startTime + ttl * 1000; // Convert TTL from seconds to milliseconds
        List<Packet> responsePackets = new ArrayList<>();
        Debugger.log("Sending monitoring packet..." + PacketMarshaller.bytesToHex(packet));

        // Step 1: Send the initial monitoring packet and wait for ACK or Response to ensure connection is live
        DatagramPacket datagramPacket = new DatagramPacket(packet, packet.length, address, UDP_PORT);
        socket.send(datagramPacket);
        Debugger.log("Monitoring packet sent. Waiting for ACK or Response...");

        boolean connectionLive = false;  // Flag to track if connection is confirmed

        while (!connectionLive) {
            socket.setSoTimeout(TIMEOUT_MS);  // Set timeout for receiving ACK or Response

            try {
                byte[] packetBuffer = new byte[1024];
                DatagramPacket recvPacket = new DatagramPacket(packetBuffer, packetBuffer.length);
                socket.receive(recvPacket);  // Wait for incoming packet

                byte[] receivedData = Arrays.copyOf(recvPacket.getData(), recvPacket.getLength());
                Debugger.log("Received packet: " + PacketMarshaller.bytesToHex(receivedData));

                Packet receivedPacket = unmarshaller.unmarshalResponse(receivedData);
                PacketType receivedType = PacketType.fromCode(receivedPacket.messageType());

                if (receivedType == PacketType.ACK) {
                    Debugger.log("ACK received. Connection is live.");
                    connectionLive = true;  // Server acknowledged the request
                } else if (receivedType == PacketType.RESPONSE) {
                    Debugger.log("Received RESPONSE packet. Connection is live.");
                    connectionLive = true;  // RESPONSE packet indicates connection is live
                    responsePackets.add(receivedPacket);  // Add the first RESPONSE packet
                    sendAckForResponse(receivedPacket);  // Send ACK for RESPONSE
                } else {
                    Debugger.log("Waiting for ACK or RESPONSE...");
                    continue;  // Wait for a valid ACK or RESPONSE
                }
            } catch (SocketTimeoutException e) {
                retries++;
                Debugger.log("Timeout waiting for ACK or RESPONSE, retrying attempt " + retries);
                if (retries >= MAX_RETRIES) {
                    throw new IOException("Failed to receive ACK or RESPONSE after " + MAX_RETRIES + " retries.");
                }
                backoffDelay(retries);  // Exponential backoff for retries
                socket.send(datagramPacket);  // Resend the monitoring packet
            }
        }

        // Step 2: Once ACK or RESPONSE is received, monitor for responses during the TTL period
        Debugger.log("Connection confirmed. Monitoring facility for " + ttl + " seconds...");
        while (System.currentTimeMillis() < ttlEndTime) {
            long remainingTime = ttlEndTime - System.currentTimeMillis();
            socket.setSoTimeout((int) Math.min(remainingTime, TIMEOUT_MS));  // Dynamically set timeout based on remaining TTL time
            try {
                byte[] packetBuffer = new byte[1024];
                DatagramPacket recvPacket = new DatagramPacket(packetBuffer, packetBuffer.length);
                socket.receive(recvPacket);  // Wait for incoming packet

                byte[] receivedData = Arrays.copyOf(recvPacket.getData(), recvPacket.getLength());
                Debugger.log("Received packet: " + PacketMarshaller.bytesToHex(receivedData));

                Packet receivedPacket = unmarshaller.unmarshalResponse(receivedData);
                PacketType receivedType = PacketType.fromCode(receivedPacket.messageType());

                if (receivedType == PacketType.ACK) {
                    Debugger.log("Received ACK packet, but we continue monitoring.");
                    continue;  // Ignore ACKs, as we're looking for RESPONSE packets
                } else if (receivedType == PacketType.RESPONSE) {
                    Debugger.log("Received RESPONSE packet " + receivedPacket.packetNumber()
                            + " of " + receivedPacket.totalPackets());
                    responsePackets.add(receivedPacket);
                    sendAckForResponse(receivedPacket);  // Send ACK for response packet
                    //TODO: Add fault tolerance here, resend request
                    continue;
                } else if (receivedType == PacketType.REQUEST_RESEND) {
                    Debugger.log("Resend packet received. This should not happen.");
                    continue;  // Handle unexpected resend requests if necessary
                }
            } catch (SocketTimeoutException e) {
                // Timeout waiting for a response
                Debugger.log("Timeout while monitoring, retrying...");
                continue;  // Retry waiting for responses
            }
        }

        // Step 3: After TTL period ends, stop monitoring and return the received responses
        Debugger.log("TTL expired. Monitoring stopped.");
        return responsePackets;  // Return all received packets within the TTL period
    }
    public List<Packet> sendPacketWithAckAndResend(byte[] packet) throws IOException {
        //This function sends a packet and waits for an acknowledgement and response.
        int retries = 0;
        long startTime = System.currentTimeMillis();
        Debugger.log("Sending packet..."+PacketMarshaller.bytesToHex(packet));
        List<Packet> responsePackets = new ArrayList<>();
        boolean isAcknowledged = false;
        while(true){
            //Firstly, send the packet.
            if (!isAcknowledged) {
                DatagramPacket DatagramPacket = new DatagramPacket(packet, packet.length, address, UDP_PORT);
                socket.send(DatagramPacket);
                Debugger.log("Packet sent");
                long elapsed = System.currentTimeMillis() - startTime;

                if (elapsed > TIMEOUT_MS) {
                    throw new IOException("Failed to send packet after " + MAX_RETRIES + " retry attempts.");
                }

                socket.setSoTimeout(TIMEOUT_MS);
            }

            try{
                byte[] packetBuffer = new byte[1024];
                DatagramPacket recvPacket = new DatagramPacket(packetBuffer, packetBuffer.length);
                socket.receive(recvPacket);
                byte[] receivedData = Arrays.copyOf(recvPacket.getData(), recvPacket.getLength());
                Debugger.log("Received packet: " + PacketMarshaller.bytesToHex(receivedData));

                Packet receivedPacket = unmarshaller.unmarshalResponse(receivedData);
                PacketType recievedType = PacketType.fromCode(receivedPacket.messageType());
                if(recievedType == PacketType.ACK){
                    Debugger.log("Recieved AcK packet");
                    isAcknowledged = true;
                    continue;
                }
                else if (recievedType == PacketType.RESPONSE){
                    Debugger.log("Received RESPONSE packet " + receivedPacket.packetNumber()
                            + " of " + receivedPacket.totalPackets());
                    responsePackets.add(receivedPacket);
                    sendAckForResponse(receivedPacket);
                    if (responsePackets.size() == receivedPacket.totalPackets()){
                        responsePackets.sort(Comparator.comparingInt(Packet::packetNumber));
                        return responsePackets;
                    }
                    continue;
                }
                else if (recievedType == PacketType.REQUEST_RESEND){
                    Debugger.log("Resend packet recieved. This should not happen.");
                    continue;
                }
            }catch (SocketTimeoutException e){
                //No packets
                retries++;
                Debugger.log("Timeout waiting for response, retrying attempt " + retries);
                if(retries>=MAX_RETRIES){
                    throw new IOException("Failed to send packet after " + MAX_RETRIES + " retry attempts.");
                }
                backoffDelay(retries);
                continue;
            }
        }

    }
    // Helper method to send back an ACK constructed from a response
    private void sendAckForResponse(Packet responsePacket) throws IOException {
        ByteBuffer payload = ByteBuffer.allocate(16 + 1);
        payload.put(PacketMarshaller.UUIDtoByteArray(responsePacket.messageId()));
        payload.put(responsePacket.packetNumber());
        byte[] combinedPayload = payload.array();
        byte[] ackPacketData = PacketMarshaller.marshalPacket(
                (byte) 0x04,  // Message Type (Request)
                (byte) 0x00,  // Packet Number (0)
                (byte) 0x01,  // Total Packets (1)
                false,          // Ack Required
                false,         // Fragment
                combinedPayload  // 0x01, Payload (StartTime, EndTime, FacilityName)
        );
        DatagramPacket ackPacket = new DatagramPacket(ackPacketData, ackPacketData.length, address, UDP_PORT);
        socket.send(ackPacket);
        Debugger.log("Sent ACK for response packet " + responsePacket.packetNumber());
    }
    // Helper method to convert byte array to hex string for easier debugging
    public static String bytesToHex(byte[] byteArray) {
        StringBuilder hexString = new StringBuilder();
        for (byte b : byteArray) {
            hexString.append(String.format("%02X", b));
        }
        return hexString.toString();
    }
    private void backoffDelay(int retryCount) {
        //Helper function to introduce backoffDelay to simulate server down time
        try {
            // Base delay: 100ms multiplied by 2^retryCount.
            long baseDelay = 100 * (long) Math.pow(2, retryCount);

            // Random factor between 0.8 and 1.2.
            double randomFactor = 0.8 + Math.random() * 0.4;

            // Apply the random factor to the base delay.
            long delay = (long) (baseDelay * randomFactor);
            Debugger.log("Delaying for " + delay + "ms before retrying...");
            Thread.sleep(delay);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
    }


}
