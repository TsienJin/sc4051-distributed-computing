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
    void testUnmarshalResponseWithValidResponse() {
        // Arrange
        PacketMarshaller marshaller = new PacketMarshaller();

        byte[] target = marshaller.marshalCreateFacilityRequest("One");
        System.out.println(marshaller.bytesToHex(target));
        marshaller.unmarshalResponse(target);
    }
}