package bookingclient;

import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.*;

class NetworkHandlerTest {

    /**
     * Test class for NetworkHandler's isAck method.
     * The isAck method determines whether a given packet is an acknowledgment
     * packet (ACK) based on its length and specific byte content.
     */

    @Test
    void testIsAckWithValidAckPacket() {
        // Arrange: Create a packet with at least 19 bytes and set byte at index 18 to ACK code.
        byte[] validAckPacket = new byte[19];
        validAckPacket[18] = PacketType.ACK.getCode();

        // Act & Assert: Verify isAck returns true for an ACK packet.
        assertTrue(NetworkHandler.isAck(validAckPacket));
    }

    @Test
    void testIsAckWithNonAckPacket() {
        // Arrange: Create a packet with at least 19 bytes and set byte at index 18 to a non-ACK value.
        byte[] nonAckPacket = new byte[19];
        nonAckPacket[18] = (byte) 0x01; // Assumes 0x01 is not the ACK code.

        // Act & Assert: Verify isAck returns false for a non-ACK packet.
        assertFalse(NetworkHandler.isAck(nonAckPacket));
    }

    @Test
    void testIsAckWithPacketTooShort() {
        // Arrange: Create a packet shorter than 18 bytes.
        byte[] shortPacket = new byte[17];

        // Act & Assert: Verify that an exception is thrown for a short packet.
        assertThrows(IllegalArgumentException.class, () -> NetworkHandler.isAck(shortPacket));
    }
}