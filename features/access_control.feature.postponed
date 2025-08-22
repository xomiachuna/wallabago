Feature: Access Control
    Background:
        Given the following actors exist:
            |actor  |credentials    |roles  |
            |admin  |valid          |admin  |
            |user   |valid          |user   |
            |rogue  |invalid        |       |

        And the following clients exist:
            |client |credentials    |
            |web    |valid          |
            |rogue  |invalid        |
    
    Scenario Outline: Authentication outcome
        Given <actor> has <cred> credentials
        And <client> has <client-cred> credentials
        When the client authenticates on behalf of the actor
        Then the client is <outcome>

        Examples:
            |actor  |cred   |client |client-cred|outcome        |
            |admin  |valid  |web    |valid      |authenticated  |
            |user   |valid  |web    |valid      |authenticated  |
            |admin  |invalid|web    |valid      |rejected       |
            |user   |invalid|web    |valid      |rejected       |
            |rogue  |invalid|web    |valid      |rejected       |
            |admin  |valid  |rogue  |invalid    |rejected       |
            |user   |valid  |rogue  |invalid    |rejected       |
            |rogue  |invalid|rogue  |invalid    |rejected       |

