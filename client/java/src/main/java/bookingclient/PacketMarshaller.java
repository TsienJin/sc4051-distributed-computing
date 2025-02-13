package bookingclient;

import java.nio.ByteBuffer;
import java.nio.charset.StandardCharsets;
import java.util.UUID;
import java.util.zip.CRC32;

public class PacketMarshaller {

    // Method to construct the packet
    public byte[] marshalPacket(byte messageType, byte packetNumber, byte totalPackets, boolean ackRequired, boolean fragment, byte methodIdentifier,byte[] payload) {
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
        short payloadLength = (short) (payload.length + 1);  // +1 byte for Method Identifier (8 bits)
        byte methodIdentifierByte = methodIdentifier;
        // 8. Checksum (32 bits, CRC32 for simplicity)
        byte[] checksum = calculateChecksum(protocolVersion, messageIdBytes, msgType, packetNo, totalPacketsByte, flags, payloadLength,methodIdentifierByte, payload);

        // Allocate a ByteBuffer with sufficient space for all the fields
        ByteBuffer buffer = ByteBuffer.allocate(1 + 16 + 1 + 1 + 1 + 1 + 2 + 1 + payload.length + 4);

        // Fill the buffer with the data
        buffer.put(protocolVersion);  // Protocol version
        buffer.put(messageIdBytes);   // Message ID (UUID)
        buffer.put(msgType);          // Message Type
        buffer.put(packetNo);         // Packet Number
        buffer.put(totalPacketsByte); // Total Packets
        buffer.put(flags);            // Flags
        buffer.putShort(payloadLength);  // Payload length (including Method Identifier)
        buffer.put(methodIdentifierByte);        // Method Identifier (Create Facility)
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
    private byte[] calculateChecksum(byte protocolVersion, byte[] messageIdBytes, byte msgType, byte packetNo, byte totalPackets, byte flags, short payloadLength,byte methodIdentifierByte, byte[] payload) {
        // Combine all parts of the packet (excluding checksum itself) into a single byte array
        ByteBuffer buffer = ByteBuffer.allocate(1 + 16 + 1 + 1 + 1 + 1 + 2 + 1 + payload.length);
        buffer.put(protocolVersion); //1
        buffer.put(messageIdBytes); //16
        buffer.put(msgType); //1
        buffer.put(packetNo); //1
        buffer.put(totalPackets); //1
        buffer.put(flags); //1
        buffer.putShort(payloadLength); // 2
        buffer.put(methodIdentifierByte);  // Add Method Identifier (e.g., Create Facility) //1
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
        return marshalPacket(
                (byte) 0x02,  // Message Type (Request)
                (byte) 0x00,  // Packet Number (0)
                (byte) 0x01,  // Total Packets (1)
                true,          // Ack Required
                false,         // Fragment
                (byte) 0x01,
                facilityBytes  // Payload (Facility Name)
        );
    }
}
