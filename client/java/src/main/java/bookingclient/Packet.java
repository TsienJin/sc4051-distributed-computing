package bookingclient;

import java.util.UUID;

/**
 * @param protocolVersion Fields from the marshalling process
 * @param status          Additional field that might be derived from the payload or header, for example, a status code. You can populate this during unmarshalling.
 */
public record Packet(byte protocolVersion, UUID messageId, byte messageType, byte packetNumber, byte totalPackets,
                     byte flags, short payloadLength, byte[] payload, byte[] checksum, int status) {

    byte[] getPayload() {
        return this.payload;
    }
}
