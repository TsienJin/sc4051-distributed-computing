import bookingclient.UserInputUtils;
import org.junit.jupiter.api.*;
import static org.junit.jupiter.api.Assertions.*;
import java.io.*;

public class UserInputUtilsTest {
    @Test
    void testGetStringInput_MultipleEmptyThenValidInput() {
        ByteArrayOutputStream outputStream = new ByteArrayOutputStream(); // Create fresh output stream for this test
        System.setOut(new PrintStream(outputStream));

        String input = "\n\nJohn Doe"; // Two empty inputs followed by valid input
        ByteArrayInputStream inputStream = new ByteArrayInputStream((input + "\n").getBytes());
        System.setIn(inputStream);

        String prompt = "Enter your name:";
        String result = UserInputUtils.getStringInput(prompt);

        assertEquals("John Doe", result);

        String[] outputs = outputStream.toString().split("\n");
        assertTrue(outputs[0].contains(prompt));
        assertTrue(outputs[1].contains("Input cannot be empty. Please enter a valid string."));
        assertTrue(outputs[2].contains(prompt));
        assertTrue(outputs[3].contains("Input cannot be empty. Please enter a valid string."));

    }
}
