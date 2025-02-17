package bookingclient;

import org.junit.jupiter.api.Test;

import java.nio.ByteBuffer;
import java.util.UUID;
import java.util.zip.CRC32;

import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertEquals;

class PacketMarshallerTest {

    /**
     * Class under test: PacketMarshaller
     * Method under test: unmarshalResponse
     * Description: The unmarshalResponse method takes a byte array representing a serialized response
     * and parses its components (protocol version, message ID, message type, etc.).
     */

    @Test
    void testUnmarshalResponseWithInvalidResponse() {
        byte[] target = new byte[]{
                (byte) 0x01, (byte) 0x18, (byte) 0x97, (byte) 0xDA, (byte) 0xFD, (byte) 0xED, (byte) 0x15, (byte) 0x11,
                (byte) 0xEF, (byte) 0x9B, (byte) 0xB0, (byte) 0x02, (byte) 0x42, (byte) 0xAC, (byte) 0x12, (byte) 0x00,
                (byte) 0x02, (byte) 0x03, (byte) 0x00, (byte) 0x01, (byte) 0x01, (byte) 0x00, (byte) 0x12,
                (byte) 0xA1,
                (byte) 0x51, (byte) 0xAA, (byte) 0x84, (byte) 0x42, (byte) 0xD6, (byte) 0x44, (byte) 0x94, (byte) 0x95,
                (byte) 0xC1, (byte) 0xBB, (byte) 0x88, (byte) 0x55, (byte) 0x4B, (byte) 0x35, (byte) 0x2E, (byte) 0x00,
                (byte) 0xC8, (byte) 0x1C, (byte) 0xA7, (byte) 0x10, (byte) 0xA9, (byte) 0x00, (byte) 0x00, (byte) 0x00,
                (byte) 0x00, (byte) 0x00, (byte) 0x00, (byte) 0x00, (byte) 0x00, (byte) 0x00, (byte) 0x00, (byte) 0x00,
                (byte) 0x00, (byte) 0x00, (byte) 0x00, (byte) 0x00, (byte) 0x00, (byte) 0x00, (byte) 0x00, (byte) 0x00,
        };
        PacketMarshaller marshaller = new PacketMarshaller();
        marshaller.unmarshalResponse(target);
    }
    @Test
    void testUnmarshalResponseWithExampleResponse() {
        // Arrange
        PacketMarshaller marshaller = new PacketMarshaller();

        // Provided byte array
        byte[] target = new byte[]{
                (byte) 0x01, (byte) 0xAF, (byte) 0x0E, (byte) 0xEC, (byte) 0xC7, (byte) 0xED, (byte) 0x06, (byte) 0x11,
                (byte) 0xEF, (byte) 0x9B, (byte) 0xB0, (byte) 0x02, (byte) 0x42, (byte) 0xAC, (byte) 0x12, (byte) 0x00,
                (byte) 0x02, (byte) 0x03, (byte) 0x00, (byte) 0x01, (byte) 0x01, (byte) 0x00, (byte) 0x29,
                (byte) 0x2D, (byte) 0x1A, (byte) 0x1D, (byte) 0xA9, (byte) 0xFD, (byte) 0xA9, (byte) 0x43, (byte) 0x08, (byte) 0xB3,
                (byte) 0x6F, (byte) 0x28, (byte) 0x3F, (byte) 0x0D, (byte) 0xCD, (byte) 0x6E, (byte) 0x46,
                (byte) 0x01, (byte) 0x90, (byte) 0x66, (byte) 0x61, (byte) 0x63, (byte) 0x69, (byte) 0x6C, (byte) 0x69, (byte) 0x74,
                (byte) 0x79, (byte) 0x20, (byte) 0x61, (byte) 0x6C, (byte) 0x72, (byte) 0x65, (byte) 0x61, (byte) 0x64,
                (byte) 0x79, (byte) 0x20, (byte) 0x65, (byte) 0x78, (byte) 0x69, (byte) 0x73, (byte) 0x74, (byte) 0x73,
                (byte) 0xCA, (byte) 0x56, (byte) 0x6D, (byte) 0x1F, (byte) 0x00, (byte) 0x00, (byte) 0x00, (byte) 0x00
                // Remaining bytes omitted for brevity
        };

        // Act
        marshaller.unmarshalResponse(target);
    }
}

