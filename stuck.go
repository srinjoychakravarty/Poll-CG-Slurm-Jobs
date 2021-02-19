package main

import (
    "bufio"
    "bytes"
    "fmt"
    "net/smtp"
    "os"
    "os/exec"
    "strings"
    "strconv"
    "time"
)

func doEvery(d time.Duration, f func(time.Time)) {
	for x := range time.Tick(d) {
		f(x)
	}
}

type stuck_job_struct struct {
    partition string
    job_name  string
    username string
    time_in_completion string
    num_of_nodes string
    node_list string
}

type node_status_struct struct {
    powercyclable bool
    running_jobs int
}

func returnStuckJobHashmap() map[string]stuck_job_struct {
    var multiline_squeue_output string = `17171178   express F[C](F)C farina.d CG       7:21      1 c0164
    17177800     short conn_054   akucyi CG       3:00      1 c0164
    17183226_25     short    Array cantosro CG    8:30:05      1 c0191
    17186523     short     bash conahan CG      43:42      1 c0191`

    stuck_job_map := make(map[string]stuck_job_struct)
    scanner := bufio.NewScanner(strings.NewReader(multiline_squeue_output))
    for scanner.Scan() {
        chunked_squeue := strings.Fields(scanner.Text())
        job_id := chunked_squeue[0]
        partition := chunked_squeue[1]
        job_name := chunked_squeue[2]
        username := chunked_squeue[3]
        time_in_completion := chunked_squeue[5]
        num_of_nodes := chunked_squeue[6]
        node_list := chunked_squeue[7]
        stuck_job_map[job_id] = stuck_job_struct{partition, job_name, username, time_in_completion, num_of_nodes, node_list}
    }
    return stuck_job_map
}

func listStuckJobNodes(stuck_job_map map[string]stuck_job_struct) []string {
    affected_nodes_slice := make([]string, len(stuck_job_map))
    var index int
    for _, value := range stuck_job_map {
        node_job_stuck_on := value.node_list
        affected_nodes_slice[index] = node_job_stuck_on
        index += 1
	}
    return affected_nodes_slice
}

func removeDuplicatesUnordered(elements []string) []string {
    encountered := map[string]bool{}    
    for v:= range elements {
        encountered[elements[v]] = true // Create a map of all unique elements.
    }
    affected_nodes_set := []string{}    // Place all keys from the map into a slice.
    for key, _ := range encountered {
        affected_nodes_set = append(affected_nodes_set, key)
    }
    return affected_nodes_set
}

func nodesPowerCyclable(affected_nodes_set []string, simulate bool) map[string]node_status_struct {
    var powercyclable bool
    var active_job_output = *new(string)
    node_status_hashmap := make(map[string]node_status_struct)
    if simulate == false { 
        fmt.Println("real mode...") // get real output from squeue command
        for _, node := range affected_nodes_set {
            lineCount := 0
            specific_node := fmt.Sprintf("--nodes=%d", node)
            cmd := exec.Command( "squeue", specific_node )   // construct `go version` command
            var out bytes.Buffer
            cmd.Stdout = &out
            cmd.Start();    // run `cmd` in background
            cmd.Wait()  // wait `cmd` until it finishes
            active_job_output = out.String()
            fmt.Printf("List of Active Jobs on Node:\n %s \n", active_job_output)
            scanner := bufio.NewScanner(strings.NewReader(active_job_output))
            for scanner.Scan() {
                lineCount++
            }
            lineCountStr := strconv.Itoa(lineCount) 
            fmt.Printf("Number of Active Jobs on Node:\n %s \n\n", lineCountStr)
            if lineCount > 0 {
                powercyclable = false
            } else {
                powercyclable = true
            }
            node_status_hashmap[node] = node_status_struct{powercyclable, lineCount}
        }
    } else {
        fmt.Println("simulation mode...")   // prepopulate fake output
        for _, node := range affected_nodes_set {
            lineCount := 0
            var multiline_nodejob_output string = `17401710     short   sys/dash   harris.s   R   1:25:16   1   c0146
            17401989_1   short   3-hem-GA   adrion.d   R   32:38     1   c0146
            17402661     short   sys/dash   stepanya   R   9:27      1   c0146`
            active_job_output = multiline_nodejob_output
            fmt.Printf("Fake Jobs shown on Node:\n %s \n\n", active_job_output)
            scanner := bufio.NewScanner(strings.NewReader(multiline_nodejob_output))
            for scanner.Scan() {
                lineCount++
            }
            actual_job_count_on_node := lineCount - 1    // golang seems to account for an extra blank line
            lineCountStr := strconv.Itoa(actual_job_count_on_node) 
            fmt.Printf("Number of Active Jobs on Node:\n %s \n\n", lineCountStr)
            if actual_job_count_on_node > 0 {
                powercyclable = false
            } else {
                powercyclable = true
            }
            node_status_hashmap[node] = node_status_struct{powercyclable, actual_job_count_on_node}
        }
    }
    return node_status_hashmap
}

func sendEmail(stuck_job_map map[string]stuck_job_struct, nodes_status_hashmap map[string]node_status_struct) {
	from := "xchakravarty@gmail.com"    // Sender data.
	password := os.Getenv("GMAIL_PASSWORD")
    fmt.Printf("Password Used for SMTP: %s \n\n", password)
	to := []string{
		"d0d65122.northeastern.onmicrosoft.com@amer.teams.ms",   // Receiver email address.
	}
	smtpHost := "smtp.gmail.com"    // smtp server configuration.    
	smtpPort := "587"
    message := "Dear SysAdmin,\n\n"
    message_buffer := bytes.NewBufferString(message)
    for key, value := range stuck_job_map {
        stuck_job_variable := fmt.Sprintf("Job ID: %s \t Job Details: %+v \n", key, value)
        message_buffer.WriteString(stuck_job_variable)
	}
    message_buffer.WriteString("\n")
    for key, value := range nodes_status_hashmap {
        list_of_active_jobs_on_node := fmt.Sprintf("Node ID: %s \t Node Job Details: %+v \n", key, value)
        message_buffer.WriteString(list_of_active_jobs_on_node)
	}
    message_buffer.WriteString("\n")
    message_buffer.WriteString("Best,\n Srinjoy Chakravarty\nGraduate Research Assistant\nResearch Computing\nNortheastern University")
    email_body := message_buffer.String()
    fmt.Println(email_body)
    msg_bytes_slice := []byte(email_body)
	auth := smtp.PlainAuth("", from, password, smtpHost)    // Authentication.
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, msg_bytes_slice)    // Sending email.
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Email sent to user!")
}

func poll_stuck_jobs(t time.Time) {
    fmt.Println("Polling new round of stuck jobs in CG...\n\n")
    var simulate bool 
    mode := os.Args[1]
    if mode == "real" {
        simulate = false
    } else if mode == "simulate" {
        simulate = true
    } else {
        simulate = true
    }
    stuck_job_map := returnStuckJobHashmap()
    affected_nodes_slice := listStuckJobNodes(stuck_job_map)
    affected_nodes_set := removeDuplicatesUnordered(affected_nodes_slice)
    nodes_status_hashmap := nodesPowerCyclable(affected_nodes_set, simulate)
    sendEmail(stuck_job_map, nodes_status_hashmap)
}

func main() {
    doEvery(30000*time.Millisecond, poll_stuck_jobs)
}

// $ module load go/1.16
// go run stuck.go simulated