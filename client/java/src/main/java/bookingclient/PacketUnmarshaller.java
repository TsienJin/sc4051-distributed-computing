package bookingclient;

import javax.sound.midi.SysexMessage;
import java.nio.ByteBuffer;
import java.nio.charset.StandardCharsets;
import java.util.Arrays;
import java.util.UUID;

public class PacketUnmarshaller {

    // Unmarshals the response into a unified Packet object.
    public static Packet unmarshalResponse(byte[] response) {
        ByteBuffer buffer = ByteBuffer.wrap(response);
        if (Debugger.isEnabled()) {
            System.out.println("[Debugger] Received response: " + bytesToHex(response));
            System.out.println("[Debugger] " + buffer.remaining());
        }

        byte protocolVersion = buffer.get();

        byte[] messageIdBytes = new byte[16];
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

        if (isChecksumValid(
                calculateChecksum(protocolVersion, messageIdBytes, msgType, packetNo, totalPackets, flags, payloadLength, payload),
                checksum)) {
            System.out.println("Checksum is valid");
        } else {
            System.out.println("Checksum is invalid");
        }

        Packet unmarshalled = createUnmarshalledPacket(protocolVersion, messageIdBytes, msgType, packetNo, totalPackets, flags, payloadLength, payload, checksum);

        // Optionally log status or error text (if status != 200)
        if (unmarshalled.status() != 200) {
            if (payloadLength > 35) {
                byte[] errorTextBytes = Arrays.copyOfRange(payload, 18, payloadLength);
                System.out.println("Error Text: " + new String(errorTextBytes, StandardCharsets.UTF_8));
            }
            System.out.println("Error Code: " + unmarshalled.status());
        } else {
            System.out.println("Status Code: " + unmarshalled.status());
        }

        if (Debugger.isEnabled()) {
            System.out.println("[Debugger] Message ID: " + fromByteArray(messageIdBytes));
            System.out.println("[Debugger] Protocol Version: " + protocolVersion);
            System.out.println("[Debugger] Message Type: " + msgType);
            System.out.println("[Debugger] Packet Number: " + packetNo);
            System.out.println("[Debugger] Total Packets: " + totalPackets);
            System.out.println("[Debugger] Flags: " + flags);
            System.out.println("[Debugger] Payload Length: " + payloadLength);
            System.out.println("[Debugger] Payload: " + bytesToHex(payload));
            System.out.println("[Debugger] Checksum: " + bytesToHex(checksum));
        }
        return unmarshalled;
    }

    // Helper to build a unified Packet from all header fields.
    private static Packet createUnmarshalledPacket(byte protocolVersion, byte[] messageIdBytes, byte msgType, byte packetNo,
                                            byte totalPackets, byte flags, short payloadLength, byte[] payload, byte[] checksum) {
        // Derive a status code from the payload.
        // For example, if the first 2 bytes of the payload represent a status code:
        int status = 200; // default (OK)
        if (payload.length >= 18) {
            ByteBuffer pb = ByteBuffer.wrap(payload);
            pb.position(16);
            status = pb.getShort();
        }
        // Create a Packet using all fields. The Packet constructor is assumed to be:
        // Packet(byte protocolVersion, UUID messageId, byte messageType, byte packetNumber, byte totalPackets,
        //        byte flags, short payloadLength, byte[] payload, byte[] checksum, int status)
        return new Packet(
                protocolVersion,
                fromByteArray(messageIdBytes),
                msgType,
                packetNo,
                totalPackets,
                flags,
                payloadLength,
                payload,
                checksum,
                status
        );
    }

    // Verify checksum (using CRC32 equivalent)
    private static boolean isChecksumValid(byte[] calculatedChecksum, byte[] receivedChecksum) {
        return ByteBuffer.wrap(calculatedChecksum).equals(ByteBuffer.wrap(receivedChecksum));
    }

    // Calculate checksum (should mirror the marshaller's implementation)
    private static byte[] calculateChecksum(byte protocolVersion, byte[] messageIdBytes, byte msgType, byte packetNo,
                                     byte totalPackets, byte flags, short payloadLength, byte[] payload) {
        ByteBuffer buffer = ByteBuffer.allocate(1 + 16 + 1 + 1 + 1 + 1 + 2 + payload.length);
        buffer.put(protocolVersion);
        buffer.put(messageIdBytes);
        buffer.put(msgType);
        buffer.put(packetNo);
        buffer.put(totalPackets);
        buffer.put(flags);
        buffer.putShort(payloadLength);
        buffer.put(payload);

        java.util.zip.CRC32 crc32 = new java.util.zip.CRC32();
        crc32.update(buffer.array());
        long checksumValue = crc32.getValue();
        ByteBuffer checksumBuffer = ByteBuffer.allocate(4);
        checksumBuffer.putInt((int) checksumValue);
        return checksumBuffer.array();
    }

    // Convert byte array to UUID
    private static UUID fromByteArray(byte[] byteArray) {
        ByteBuffer buffer = ByteBuffer.wrap(byteArray);
        long mostSigBits = buffer.getLong();
        long leastSigBits = buffer.getLong();
        return new UUID(mostSigBits, leastSigBits);
    }

    // Helper method to convert byte array to hex string for debugging
    public static String bytesToHex(byte[] byteArray) {
        StringBuilder hexString = new StringBuilder();
        for (byte b : byteArray) {
            hexString.append(String.format("%02X", b));
        }
        return hexString.toString();
    }
    public static String monitoredPayloadBytesToString(byte[] payload) {
        if (payload.length > 35) {
            byte[] LogBytes = Arrays.copyOfRange(payload, 18, payload.length);
           String log =new String(LogBytes, StandardCharsets.UTF_8);
           return log;
        }
        else{
            return null;
        }
    }
}
