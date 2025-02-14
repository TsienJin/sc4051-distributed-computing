package bookingclient;

import java.nio.ByteBuffer;
import java.nio.charset.StandardCharsets;
import java.util.UUID;
import java.util.zip.CRC32;

public class PacketMarshaller {

    // Method to construct the packet
    public byte[] marshalPacket(byte messageType, byte packetNumber, byte totalPackets, boolean ackRequired, boolean fragment, byte[] payload) {
        // 1. Protocol version (8 bits)
        byte protocolVersion = 0x01;

        // 2. Message ID (UUID, 128 bits / 16 bytes)
        UUID messageId = UUID.randomUUID();
        System.out.println("Message ID: " + messageId);
        byte[] messageIdBytes = toByteArray(messageId);

        // 3. Message Type (8 bits)
        byte msgType = messageType;

        // 4. Packet Number (8 bits)
        byte packetNo = packetNumber;

        // 5. Total Packets (8 bits)
        byte totalPacketsByte = totalPackets;

        // 6. Flags (8 bits, for example, Ack Required = 1, Fragment = 0)
        byte flags = 0;
        if (ackRequired) {
            flags |= (1 << 0);  // Set LSB for Ack Required
        }
        if (fragment) {
            flags |= (1 << 1);  // Set 2nd LSB for Fragment
        }

        // 7. Payload Length (16 bits, size of the payload)
        short payloadLength = (short) (payload.length);  // +1 byte for Method Identifier (8 bits)
        // 8. Checksum (32 bits, CRC32 for simplicity)
        byte[] checksum = calculateChecksum(protocolVersion, messageIdBytes, msgType, packetNo, totalPacketsByte, flags, payloadLength, payload);

        // Allocate a ByteBuffer with sufficient space for all the fields
        ByteBuffer buffer = ByteBuffer.allocate(1 + 16 + 1 + 1 + 1 + 1 + 2 + payload.length + 4);

        // Fill the buffer with the data
        buffer.put(protocolVersion);  // Protocol version
        buffer.put(messageIdBytes);   // Message ID (UUID)
        buffer.put(msgType);          // Message Type
        buffer.put(packetNo);         // Packet Number
        buffer.put(totalPacketsByte); // Total Packets
        buffer.put(flags);            // Flags
        buffer.putShort(payloadLength);  // Payload length (including Method Identifier)
        buffer.put(payload);          // Payload (Facility Name)
        buffer.put(checksum);         // Checksum

        // Return the final packet as a byte array
        return buffer.array();
    }

    // Convert UUID to byte array (16 bytes)
    private byte[] toByteArray(UUID uuid) {
        ByteBuffer buffer = ByteBuffer.wrap(new byte[16]);
        buffer.putLong(uuid.getMostSignificantBits());
        buffer.putLong(uuid.getLeastSignificantBits());
        return buffer.array();
    }

    // Calculate checksum (simple CRC32 for the whole packet)
    private byte[] calculateChecksum(byte protocolVersion, byte[] messageIdBytes, byte msgType, byte packetNo, byte totalPackets, byte flags, short payloadLength, byte[] payload) {
        // Combine all parts of the packet (excluding checksum itself) into a single byte array
        ByteBuffer buffer = ByteBuffer.allocate(1 + 16 + 1 + 1 + 1 + 1 + 2 + payload.length);
        buffer.put(protocolVersion); //1
        buffer.put(messageIdBytes); //16
        buffer.put(msgType); //1
        buffer.put(packetNo); //1
        buffer.put(totalPackets); //1
        buffer.put(flags); //1
        buffer.putShort(payloadLength); // 2
        buffer.put(payload); //payload.length

        // Use CRC32 for checksum calculation
        CRC32 crc32 = new CRC32();
        crc32.update(buffer.array());

        // Return the checksum as a 4-byte array
        long checksumValue = crc32.getValue();
        ByteBuffer checksumBuffer = ByteBuffer.allocate(4);
        checksumBuffer.putInt((int) checksumValue);
        return checksumBuffer.array();
    }
    public byte[] unmarshalResponse(byte[] response) {

        ByteBuffer buffer = ByteBuffer.wrap(response);
        if(Debugger.isEnabled()){
            System.out.println("Received response: " + bytesToHex(response));
            System.out.println(buffer.remaining());
        }
        byte[] messageIdBytes = new byte[16];
        byte protocolVersion = buffer.get();
        buffer.get(messageIdBytes);
        byte msgType = buffer.get();
        byte packetNo = buffer.get();
        byte totalPackets = buffer.get();
        byte flags = buffer.get();
        short payloadLength = buffer.getShort();
        byte[] payload = new byte[payloadLength];
        buffer.get(payload);
        byte[] checksum = new byte[4];
        buffer.get(checksum);
        if(isChecksumValid(
                calculateChecksum(protocolVersion, messageIdBytes, msgType, packetNo, totalPackets, flags, payloadLength, payload),
                checksum)){
            System.out.println("Checksum is valid");
        }

        System.out.println("Payload String: " + new String(payload, StandardCharsets.UTF_8));

        if (Debugger.isEnabled()){
            System.out.println("Message ID: " + fromByteArray(messageIdBytes));
            System.out.println("Protocol Version: " + protocolVersion);
            System.out.println("Message Type: " + msgType);
            System.out.println("Packet Number: " + packetNo);
            System.out.println("Total Packets: " + totalPackets);
            System.out.println("Flags: " + flags);
            System.out.println("Payload Length: " + payloadLength);
            System.out.println("Payload: " + bytesToHex(payload));
            System.out.println("Checksum: " + bytesToHex(checksum));
        }

        return response;
    }
    // Verify checksum
    private boolean isChecksumValid(byte[] calculatedChecksum, byte[] receivedChecksum) {
        return ByteBuffer.wrap(calculatedChecksum).equals(ByteBuffer.wrap(receivedChecksum));
    }

    // Convert byte array to UUID
    private UUID fromByteArray(byte[] byteArray) {
        ByteBuffer buffer = ByteBuffer.wrap(byteArray);
        long mostSigBits = buffer.getLong();
        long leastSigBits = buffer.getLong();
        return new UUID(mostSigBits, leastSigBits);
    }

    // Helper method to convert byte array to hex string for easier debugging
    public static String bytesToHex(byte[] byteArray) {
        StringBuilder hexString = new StringBuilder();
        for (byte b : byteArray) {
            hexString.append(String.format("%02X", b));
        }
        return hexString.toString();
    }

    public byte[] marshalCreateFacilityRequest(String facility) {
        // Facility Name as the payload
        byte[] facilityBytes = facility.getBytes(StandardCharsets.UTF_8);
        byte methodIdentifier = 0x01;
        ByteBuffer buffer = ByteBuffer.allocate(1 + facilityBytes.length);
        buffer.put(methodIdentifier);
        buffer.put(facilityBytes);
        facilityBytes = buffer.array();
        return marshalPacket(
                (byte) 0x02,  // Message Type (Request)
                (byte) 0x00,  // Packet Number (0)
                (byte) 0x01,  // Total Packets (1)
                true,          // Ack Required
                false,         // Fragment
                facilityBytes  // 0x01, Payload (Facility Name)
        );
    }
    public byte[] marshalDeleteFacilityRequest(String facility) {
        // Facility Name as the payload
        byte[] facilityBytes = facility.getBytes(StandardCharsets.UTF_8);
        byte methodIdentifier = 0x04;
        ByteBuffer buffer = ByteBuffer.allocate(1 + facilityBytes.length);
        buffer.put(methodIdentifier);
        buffer.put(facilityBytes);
        facilityBytes = buffer.array();
        return marshalPacket(
                (byte) 0x02,  // Message Type (Request)
                (byte) 0x00,  // Packet Number (0)
                (byte) 0x01,  // Total Packets (1)
                true,          // Ack Required
                false,         // Fragment
                facilityBytes  // 0x01, Payload (Facility Name)
        );
    }

    public byte[] marshalQueryFacilityRequest(String facility, int numberOfDays) {
        // Facility Name as the payload
        byte[] facilityBytes = facility.getBytes(StandardCharsets.UTF_8);
        String fac = bytesToHex(facilityBytes);
//        System.out.println("facility: " + fac);

        ByteBuffer tmpBuffer = ByteBuffer.allocate(4); // Allocate 4 bytes (size of an int)
        tmpBuffer.putInt(numberOfDays); // Put the int into the buffer
        byte[] numberOfDaysBytes = tmpBuffer.array(); // Retrieve the byte array

        byte methodIdentifier = 0x02;
        ByteBuffer buffer = ByteBuffer.allocate(1 + facilityBytes.length + numberOfDaysBytes.length);
        buffer.put(methodIdentifier);
        buffer.put(facilityBytes);
        buffer.put(numberOfDaysBytes);

        byte[] combinedPayload = buffer.array();
        String x = bytesToHex(combinedPayload);
//        System.out.println("Combined Payload: " + x);
//        System.out.println("payload length = " + combinedPayload.length);

//        facilityBytes = buffer.array();
        return marshalPacket(
                (byte) 0x02,  // Message Type (Request)
                (byte) 0x00,  // Packet Number (0)
                (byte) 0x01,  // Total Packets (1)
                true,          // Ack Required
                false,         // Fragment
                combinedPayload  // 0x01, Payload (Facility Name)
        );
    }
}
